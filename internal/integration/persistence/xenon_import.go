package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/importkind"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/agecrypt"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

const unreadableExportMsg = " cannot be read: "

func (x *xenonExportPersistence) ReadXenonExport(ctx context.Context, input entity.ImportInput) (*entity.XenonImport, error) {
	logger.Debug(ctx, "Started", input)

	var read *entity.XenonImport
	var err error
	switch input.Kind {
	case importkind.XenonRequest:
		read, err = readXenonRequest(input.Document())
	case importkind.XenonEnvironment:
		read, err = readXenonEnvironment(input.Document())
	case importkind.XenonBundle:
		read, err = readXenonBundle(input.Document())
	case importkind.XenonZip:
		read, err = readXenonZip(input)
	default:
		read, err = readXenonWorkspace(input)
	}
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", input, map[string]any{"result": *read})
	return read, nil
}

func (x *xenonExportPersistence) Decrypt(ctx context.Context, armored, passphrase string) (*entity.SecretBundle, error) {
	logger.Debug(ctx, "Started", map[string]any{"bytes": len(armored)})

	plaintext, err := agecrypt.DecryptArmoredWithPassphrase([]byte(armored), passphrase)
	if err != nil {
		return nil, errs.UnauthorizedError("the passphrase does not open the export's secrets; try again or import without secrets", "IMP-08401")
	}
	var stored model.ExportSecrets
	if err := json.Unmarshal(plaintext, &stored); err != nil {
		return nil, errs.UnprocessableEntityError("the export's secrets cannot be read", "IMP-44422")
	}
	if stored.IsNewer() {
		return nil, errs.UnprocessableEntityError("the export's secrets were written by a newer Xenon; update Xenon to import them", "IMP-45422")
	}
	bundle := stored.ToEntity()

	logger.Debug(ctx, "Finished", bundle)
	return &bundle, nil
}

func readXenonRequest(document entity.ImportDocument) (*entity.XenonImport, error) {
	parsed, header, err := parseExportDocument(document)
	if err != nil {
		return nil, err
	}
	var stored model.Request
	if err := parsed.Decode(&stored); err != nil {
		return nil, unreadableExport(document.Path, err)
	}
	root := entity.NewRequestImportNode(stored.ToEntity())
	return &entity.XenonImport{Root: &root, Secrets: header.Secrets}, nil
}

func readXenonEnvironment(document entity.ImportDocument) (*entity.XenonImport, error) {
	parsed, header, err := parseExportDocument(document)
	if err != nil {
		return nil, err
	}
	var stored model.Environment
	if err := parsed.Decode(&stored); err != nil {
		return nil, unreadableExport(document.Path, err)
	}
	environment := stored.ToEntity(filepath.Base(document.Path))
	return &entity.XenonImport{Environment: &environment, Secrets: header.Secrets}, nil
}

func readXenonBundle(document entity.ImportDocument) (*entity.XenonImport, error) {
	parsed, _, err := parseExportDocument(document)
	if err != nil {
		return nil, err
	}
	var bundle model.ExportBundle
	if err := parsed.Decode(&bundle); err != nil {
		return nil, unreadableExport(document.Path, err)
	}
	root, err := bundleImportNode(document.Path, &bundle.Item)
	if err != nil {
		return nil, err
	}
	return &entity.XenonImport{Root: root, Secrets: bundle.Secrets}, nil
}

func bundleImportNode(name string, item *yaml.Node) (*entity.ImportNode, error) {
	var header model.XenonHeader
	if err := item.Decode(&header); err != nil {
		return nil, unreadableExport(name, err)
	}
	kind := itemkind.GetItemKind(header.Kind)
	if kind == itemkind.Request {
		var stored model.Request
		if err := item.Decode(&stored); err != nil {
			return nil, unreadableExport(name, err)
		}
		node := entity.NewRequestImportNode(stored.ToEntity())
		return &node, nil
	}
	if kind != itemkind.Collection && kind != itemkind.Folder {
		return nil, errs.UnprocessableEntityError(name+" holds an item that is not a collection, folder or request", "IMP-46422")
	}
	var stored model.Collection
	var children model.ExportBundleChildren
	if err := item.Decode(&stored); err != nil {
		return nil, unreadableExport(name, err)
	}
	if err := item.Decode(&children); err != nil {
		return nil, unreadableExport(name, err)
	}
	nodes := make([]entity.ImportNode, 0, len(children.Children))
	for index := range children.Children {
		child, err := bundleImportNode(name, &children.Children[index])
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, *child)
	}
	node := entity.NewContainerImportNode(stored.ToContainerEntity(stored.Id, kind), nodes)
	return &node, nil
}

func readXenonZip(input entity.ImportInput) (*entity.XenonImport, error) {
	dir, err := exportedDir(input)
	if err != nil {
		return nil, err
	}
	root, err := containerImportNode(dir, itemkind.GetItemKind(input.Probe.XenonRoot))
	if err != nil {
		return nil, err
	}
	secrets, err := exportSecretsOf(input.Root)
	if err != nil {
		return nil, err
	}
	return &entity.XenonImport{Root: root, Secrets: secrets}, nil
}

func containerImportNode(dir string, kind enums.ItemKind) (*entity.ImportNode, error) {
	stored, reason := readCollection(filepath.Join(dir, markerFile(kind)), kind)
	if reason != "" {
		return nil, errs.UnprocessableEntityError(reason, "IMP-47422")
	}
	container := stored.ToContainerEntity(stored.Id, kind)
	listing := readEntries(dir)
	folders, _ := (&treeScan{}).classifyDirectories(dir, listing, folderMarker)
	requests := requestFiles(listing)
	reads := readRequests(dir, requests)
	children := make([]entity.ImportNode, 0)
	for _, name := range container.ArrangedChildren(append(append([]string{}, folders...), requests...)) {
		if slices.Contains(folders, name) {
			child, err := containerImportNode(filepath.Join(dir, name), itemkind.Folder)
			if err != nil {
				return nil, err
			}
			children = append(children, *child)
			continue
		}
		if reads[name].reason != "" {
			return nil, errs.UnprocessableEntityError(reads[name].reason, "IMP-47422")
		}
		children = append(children, entity.NewRequestImportNode(reads[name].stored.ToEntity()))
	}
	node := entity.NewContainerImportNode(container, children)
	return &node, nil
}

func readXenonWorkspace(input entity.ImportInput) (*entity.XenonImport, error) {
	folder := input.Root
	secrets := ""
	if input.Kind == importkind.XenonWorkspaceZip {
		dir, err := exportedDir(input)
		if err != nil {
			return nil, err
		}
		if secrets, err = exportSecretsOf(input.Root); err != nil {
			return nil, err
		}
		folder = dir
	}
	_, stored, err := readWorkspace(folder)
	if err != nil {
		return nil, err
	}
	index, err := scanTree(folder, nil)
	if err != nil {
		return nil, err
	}
	environments, err := readEnvironments(folder)
	if err != nil {
		return nil, err
	}
	workspace := entity.NewImportWorkspace(folder, stored.Name, index.Tree.Collections(), environments)
	return &entity.XenonImport{Workspace: &workspace, Secrets: secrets}, nil
}

func parseExportDocument(document entity.ImportDocument) (*yaml.Node, *model.XenonHeader, error) {
	parsed, err := yamlnode.Parse(document.Data)
	if err != nil {
		return nil, nil, unreadableExport(document.Path, err)
	}
	var header model.XenonHeader
	if err := parsed.Decode(&header); err != nil {
		return nil, nil, unreadableExport(document.Path, err)
	}
	return parsed, &header, nil
}

func exportedDir(input entity.ImportInput) (string, error) {
	dirs := make([]string, 0, 1)
	for _, entry := range input.Probe.Entries {
		if dir, _, nested := strings.Cut(entry, "/"); nested && !strings.HasPrefix(dir, ".") && !slices.Contains(dirs, dir) {
			dirs = append(dirs, dir)
		}
	}
	if len(dirs) != 1 {
		return "", errs.UnprocessableEntityError(input.Probe.Name+" must hold exactly one exported folder next to "+exportManifestFile, "IMP-48422")
	}
	return filepath.Join(input.Root, dirs[0]), nil
}

func exportSecretsOf(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, exportSecretsFile))
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", errs.FileError(err, "IMP-49500")
	}
	return string(data), nil
}

func unreadableExport(name string, err error) error {
	return errs.UnprocessableEntityError(filepath.Base(name)+unreadableExportMsg+err.Error(), "IMP-47422")
}
