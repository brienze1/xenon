package persistence

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/internal/integration/persistence/yamlnode"
	"github.com/brienze1/xenon/pkg/idutils"
)

const (
	collectionMarker = "collection.yaml"
	folderMarker     = "folder.yaml"
	requestExtension = ".yaml"
	errorIdPrefix    = "error:"
	duplicateWarning = "another item has the same id, so this one uses a new id that is written when it is saved"
	missingIdWarning = "the file has no id, so it gets one that is written when it is saved"
)

type treeScan struct {
	previous     map[string]string
	seen         map[string]string
	items        map[string]indexedItem
	reidentified map[string]string
	root         string
}

func scanTree(root string, previous map[string]string) (*rootIndex, error) {
	_, workspace, err := readWorkspace(root)
	if err != nil {
		return nil, err
	}
	scan := &treeScan{
		root:         root,
		previous:     previous,
		seen:         map[string]string{},
		items:        map[string]indexedItem{},
		reidentified: map[string]string{},
	}
	dir := filepath.Join(root, collectionsDir)
	listing := readEntries(dir)
	marked, unmarked := scan.classifyDirectories(dir, listing, collectionMarker)
	collections := make([]entity.TreeNode, 0, len(marked))
	for _, name := range effectiveOrder(workspace.Order, marked) {
		collections = append(collections, scan.container(filepath.Join(dir, name), name, itemkind.Collection))
	}
	collections = append(collections, scan.ignoredNodes(dir, unmarked, collectionMarker)...)
	return &rootIndex{tree: entity.NewTree(collections, entryNames(listing)), items: scan.items, reidentified: scan.reidentified}, nil
}

func (s *treeScan) container(dir, fileName string, kind enums.ItemKind) entity.TreeNode {
	stored, reason := readCollection(filepath.Join(dir, markerFile(kind)), kind)
	if reason != "" {
		return s.errorNode(dir, fileName, reason, false)
	}
	id, warning := s.identify(stored.Id, dir)
	listing := readEntries(dir)
	folders, unmarked := s.classifyDirectories(dir, listing, folderMarker)
	requests := requestFiles(listing)
	node := entity.TreeNode{ID: id, Kind: kind, Name: stored.Name, Warning: warning, FileName: fileName, Entries: entryNames(listing)}
	for _, name := range effectiveOrder(stored.Order, append(append([]string{}, folders...), requests...)) {
		if contains(folders, name) {
			node.Children = append(node.Children, s.container(filepath.Join(dir, name), name, itemkind.Folder))
			continue
		}
		node.Children = append(node.Children, s.request(filepath.Join(dir, name), name))
	}
	node.Children = append(node.Children, s.ignoredNodes(dir, unmarked, folderMarker)...)
	s.items[id] = indexedItem{path: dir, kind: kind}
	return node
}

func (s *treeScan) request(path, fileName string) entity.TreeNode {
	stored, reason := readRequest(path)
	if reason != "" {
		return s.errorNode(path, fileName, reason, false)
	}
	id, warning := s.identify(stored.Id, path)
	request := stored.ToEntity()
	request.ID = id
	node := entity.NewRequestNode(request, fileName)
	node.Warning = warning
	s.items[id] = indexedItem{path: path, kind: itemkind.Request}
	return node
}

func (s *treeScan) identify(storedId, path string) (string, string) {
	if storedId != "" && s.seen[storedId] == "" {
		s.seen[storedId] = path
		return storedId, ""
	}
	id, known := s.previous[path]
	if !known || s.seen[id] != "" {
		id = idutils.NewUlid()
	}
	s.seen[id] = path
	s.reidentified[path] = id
	if storedId == "" {
		return id, missingIdWarning
	}
	return id, duplicateWarning
}

func (s *treeScan) errorNode(path, fileName, reason string, ignored bool) entity.TreeNode {
	relative, _ := filepath.Rel(s.root, path)
	return entity.TreeNode{
		ID:       errorIdPrefix + filepath.ToSlash(relative),
		Kind:     itemkind.Error,
		Name:     fileName,
		FileName: fileName,
		Error:    reason,
		Ignored:  ignored,
	}
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

func effectiveOrder(listed, present []string) []string {
	order := make([]string, 0, len(present))
	for _, name := range listed {
		if contains(present, name) && !contains(order, name) {
			order = append(order, name)
		}
	}
	rest := make([]string, 0, len(present))
	for _, name := range present {
		if !contains(order, name) {
			rest = append(rest, name)
		}
	}
	sort.Strings(rest)
	return append(order, rest...)
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

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
