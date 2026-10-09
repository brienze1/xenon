package persistence

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

const (
	collectionMarker = "collection.yaml"
	folderMarker     = "folder.yaml"
	requestExtension = ".yaml"
)

type treeScan struct {
	identities *entity.ItemIdentities
	items      map[string]model.IndexedItem
	root       string
}

func scanTree(root string, previous map[string]string) (*model.TreeIndex, error) {
	_, stored, err := readWorkspace(root)
	if err != nil {
		return nil, err
	}
	workspace := stored.ToEntity()
	scan := &treeScan{
		root:       root,
		identities: entity.NewItemIdentities(previous),
		items:      map[string]model.IndexedItem{},
	}
	dir := filepath.Join(root, collectionsDir)
	listing := readEntries(dir)
	marked, unmarked := scan.classifyDirectories(dir, listing, collectionMarker)
	collections := make([]entity.TreeNode, 0, len(marked))
	for _, name := range workspace.ArrangedCollections(marked) {
		collections = append(collections, scan.container(filepath.Join(dir, name), name, itemkind.Collection))
	}
	collections = append(collections, scan.ignoredNodes(dir, unmarked, collectionMarker)...)
	return &model.TreeIndex{Tree: entity.NewTree(collections, entryNames(listing)), Items: scan.items, Reidentified: scan.identities.Reidentified()}, nil
}

func (s *treeScan) container(dir, fileName string, kind enums.ItemKind) entity.TreeNode {
	stored, reason := readCollection(filepath.Join(dir, markerFile(kind)), kind)
	if reason != "" {
		return s.errorNode(dir, fileName, reason, false)
	}
	id, warning := s.identities.Identify(stored.Id, dir, idutils.NewUlid)
	container := stored.ToContainerEntity(id, kind)
	listing := readEntries(dir)
	folders, unmarked := s.classifyDirectories(dir, listing, folderMarker)
	requests := requestFiles(listing)
	node := entity.NewContainerNode(container, fileName, entryNames(listing)).WithWarning(warning)
	for _, name := range container.ArrangedChildren(append(append([]string{}, folders...), requests...)) {
		if slices.Contains(folders, name) {
			node.Children = append(node.Children, s.container(filepath.Join(dir, name), name, itemkind.Folder))
			continue
		}
		node.Children = append(node.Children, s.request(filepath.Join(dir, name), name))
	}
	node.Children = append(node.Children, s.ignoredNodes(dir, unmarked, folderMarker)...)
	s.items[id] = model.IndexedItem{Path: dir, Kind: kind}
	return node
}

func (s *treeScan) request(path, fileName string) entity.TreeNode {
	stored, reason := readRequest(path)
	if reason != "" {
		return s.errorNode(path, fileName, reason, false)
	}
	id, warning := s.identities.Identify(stored.Id, path, idutils.NewUlid)
	request := stored.ToStoredEntity(id, "")
	s.items[id] = model.IndexedItem{Path: path, Kind: itemkind.Request}
	return entity.NewRequestNode(request, fileName).WithWarning(warning)
}

func (s *treeScan) errorNode(path, fileName, reason string, ignored bool) entity.TreeNode {
	relative, _ := filepath.Rel(s.root, path)
	return entity.NewErrorNode(filepath.ToSlash(relative), fileName, reason, ignored)
}

func (s *treeScan) ignoredNodes(dir string, names []string, marker string) []entity.TreeNode {
	nodes := make([]entity.TreeNode, 0, len(names))
	for _, name := range names {
		reason := fmt.Sprintf("%s is ignored because it has no %s", name, marker)
		nodes = append(nodes, s.errorNode(filepath.Join(dir, name), name, reason, true))
	}
	return nodes
}

func (s *treeScan) classifyDirectories(dir string, listing []fs.DirEntry, marker string) ([]string, []string) {
	marked, unmarked := []string{}, []string{}
	for _, entry := range listing {
		if !entry.IsDir() || isHidden(entry) {
			continue
		}
		if isRegularFile(filepath.Join(dir, entry.Name(), marker)) {
			marked = append(marked, entry.Name())
			continue
		}
		unmarked = append(unmarked, entry.Name())
	}
	return marked, unmarked
}

func requestFiles(listing []fs.DirEntry) []string {
	names := []string{}
	for _, entry := range listing {
		name := entry.Name()
		if entry.Type().IsRegular() && !isHidden(entry) && strings.EqualFold(filepath.Ext(name), requestExtension) && !isMarker(name) {
			names = append(names, name)
		}
	}
	return names
}

func readCollection(path string, kind enums.ItemKind) (*model.Collection, string) {
	var stored model.Collection
	if reason := readDocument(path, kind, &stored.Xenon, &stored.Kind, &stored); reason != "" {
		return nil, reason
	}
	return &stored, ""
}

func readRequest(path string) (*model.Request, string) {
	var stored model.Request
	if reason := readDocument(path, itemkind.Request, &stored.Xenon, &stored.Kind, &stored); reason != "" {
		return nil, reason
	}
	return &stored, ""
}

func readDocument(path string, kind enums.ItemKind, version *int, storedKind *string, target any) string {
	name := filepath.Base(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return name + " cannot be read: " + err.Error()
	}
	document, err := yamlnode.Parse(data)
	if err != nil {
		return name + " cannot be read: " + err.Error()
	}
	if err := document.Decode(target); err != nil {
		return name + " cannot be read: " + err.Error()
	}
	if *version > latestFormatVersion {
		return fmt.Sprintf("%s was written by a newer Xenon (format %d)", name, *version)
	}
	if *storedKind != "" && !kind.Equals(*storedKind) {
		return fmt.Sprintf("%s is a %s file, expected a %s", name, *storedKind, kind)
	}
	return ""
}

func readEntries(dir string) []fs.DirEntry {
	listing, _ := os.ReadDir(dir)
	return listing
}

func entryNames(listing []fs.DirEntry) []string {
	names := make([]string, 0, len(listing))
	for _, entry := range listing {
		names = append(names, entry.Name())
	}
	return names
}

func markerFile(kind enums.ItemKind) string {
	if kind == itemkind.Collection {
		return collectionMarker
	}
	return folderMarker
}

func isMarker(name string) bool {
	return strings.EqualFold(name, collectionMarker) || strings.EqualFold(name, folderMarker)
}

func isHidden(entry fs.DirEntry) bool {
	return strings.HasPrefix(entry.Name(), ".") || entry.Type()&fs.ModeSymlink != 0
}

func isRegularFile(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular()
}
