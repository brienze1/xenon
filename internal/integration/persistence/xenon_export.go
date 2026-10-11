package persistence

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/exportitemkind"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/agecrypt"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

const (
	exportManifestFile = "xenon-export.yaml"
	exportSecretsFile  = "xenon-secrets.age"
	exportTempSuffix   = ".xenon-tmp"
	localStateDir      = ".xenon"
	gitDir             = ".git"
	exporterPrefix     = "Xenon/"
	secretsBlockHeader = model.SecretsKey + ": |\n"
	secretsBlockIndent = "  "
)

type xenonExportPersistence struct {
	index adapter.WorkspaceIndexCache
}

func NewXenonExportPersistence(index adapter.WorkspaceIndexCache) adapter.XenonExportPersistence {
	return &xenonExportPersistence{index: index}
}

func (x *xenonExportPersistence) WriteXenonExport(ctx context.Context, root string, export entity.XenonExport) ([]entity.ExportEntry, error) {
	logger.Debug(ctx, "Started", export, map[string]any{"root": root})

	var entries []entity.ExportEntry
	var err error
	switch {
	case export.Kind == exportitemkind.Workspace:
		entries, err = archiveEntries(root, export)
	case export.Kind == exportitemkind.Environment:
		entries, err = environmentEntries(root, export)
	default:
		entries, err = x.itemEntries(ctx, root, export)
	}
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", export, map[string]any{"root": root, "entries": len(entries)})
	return entries, nil
}

func (x *xenonExportPersistence) Encrypt(ctx context.Context, secrets entity.SecretBundle, passphrase string) (string, error) {
	logger.Debug(ctx, "Started", secrets)

	plaintext, err := json.Marshal(model.NewExportSecrets(secrets))
	if err != nil {
		return "", errs.InternalServerError(err, "EXP-14500", "the secrets could not be encoded")
	}
	armored, err := agecrypt.EncryptArmoredWithPassphrase(plaintext, passphrase, properties.Properties().Vault.ScryptLogN)
	if err != nil {
		return "", errs.InternalServerError(err, "EXP-15500", "the secrets could not be encrypted with the passphrase")
	}

	logger.Debug(ctx, "Finished", secrets, map[string]any{"bytes": len(armored)})
	return string(armored), nil
}

func (x *xenonExportPersistence) itemEntries(ctx context.Context, root string, export entity.XenonExport) ([]entity.ExportEntry, error) {
	x.index.Lock(ctx)
	defer x.index.Unlock(ctx)

	index, item, found, err := x.index.Resolve(ctx, root, export.ItemID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errs.NotFoundError("the exported item does not exist", "EXP-16404")
	}
	switch {
	case export.IsArchive():
		return archiveEntries(item.Path, export)
	case item.IsRequest():
		return fileEntries(item.Path, export)
	default:
		return bundleEntries(index, export)
	}
}

func archiveEntries(dir string, export entity.XenonExport) ([]entity.ExportEntry, error) {
	manifest := model.NewExportManifest(export.Kind.String(), exportedAt(), exportedBy(), export.CurrentValuesAsInitial, export.HasSecrets())
	data, err := yamlnode.Encode(manifest.Document())
	if err != nil {
		return nil, errs.InternalServerError(err, "EXP-17500", "the export manifest could not be encoded")
	}
	entries := []entity.ExportEntry{entity.NewExportEntry(exportManifestFile, data)}
	files, err := exportedFiles(dir, filepath.Base(dir), export)
	if err != nil {
		return nil, err
	}
	entries = append(entries, files...)
	if export.HasSecrets() {
		entries = append(entries, entity.NewExportEntry(exportSecretsFile, []byte(export.Secrets)))
	}
	return entries, nil
}

func exportedFiles(dir, prefix string, export entity.XenonExport) ([]entity.ExportEntry, error) {
	entries := make([]entity.ExportEntry, 0)
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && path != dir && isLocalOnlyDir(entry.Name()) {
			return filepath.SkipDir
		}
		if !entry.Type().IsRegular() || strings.HasSuffix(entry.Name(), exportTempSuffix) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		entries = append(entries, entity.NewExportEntry(prefix+"/"+filepath.ToSlash(relative), withExportedVariables(path, data, export)))
		return nil
	})
	if err != nil {
		return nil, errs.FileError(err, "EXP-18500")
	}
	return entries, nil
}

func environmentEntries(root string, export entity.XenonExport) ([]entity.ExportEntry, error) {
	environment, err := findEnvironment(root, export.ItemID)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(root, environmentsDir, environment.FileName)
	return fileEntries(path, export)
}

func fileEntries(path string, export entity.XenonExport) ([]entity.ExportEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errs.FileError(err, "EXP-19500")
	}
	data = withSecretsBlock(withExportedVariables(path, data, export), export.Secrets)
	return []entity.ExportEntry{entity.NewExportEntry(filepath.Base(path), data)}, nil
}

func bundleEntries(index *entity.TreeIndex, export entity.XenonExport) ([]entity.ExportEntry, error) {
	item, err := bundleItem(index, *export.Node, export)
	if err != nil {
		return nil, err
	}
	bundle := model.NewExportBundle(export.Kind.String(), exportedAt(), exportedBy(), item, export.Secrets)
	data, err := yamlnode.Encode(bundle.Document())
	if err != nil {
		return nil, errs.InternalServerError(err, "EXP-20500", "the bundle could not be encoded")
	}
	return []entity.ExportEntry{entity.NewExportEntry(export.Node.Name(), data)}, nil
}

func bundleItem(index *entity.TreeIndex, node entity.ExportNode, export entity.XenonExport) (*yaml.Node, error) {
	id := node.Container.ID
	if node.IsRequest() {
		id = node.Request.ID
	}
	indexed, _ := index.Item(id)
	path := indexed.Path
	if !node.IsRequest() {
		path = filepath.Join(indexed.Path, markerFile(indexed.Kind))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errs.FileError(err, "EXP-21500")
	}
	document, err := yamlnode.Parse(withExportedVariables(path, data, export))
	if err != nil {
		return nil, errs.UnprocessableEntityError(filepath.Base(path)+" cannot be read: "+err.Error(), "EXP-22422")
	}
	item := yamlnode.Root(document)
	if node.IsRequest() {
		return item, nil
	}
	children := make([]*yaml.Node, 0, len(node.Children))
	for _, child := range node.Children {
		childItem, err := bundleItem(index, child, export)
		if err != nil {
			return nil, err
		}
		children = append(children, childItem)
	}
	return model.WithChildren(item, children), nil
}

func withExportedVariables(path string, data []byte, export entity.XenonExport) []byte {
	if !export.HasVariables() || !strings.EqualFold(filepath.Ext(path), requestExtension) {
		return data
	}
	document, err := yamlnode.Parse(data)
	if err != nil {
		return data
	}
	var header model.XenonHeader
	if err := document.Decode(&header); err != nil {
		return data
	}
	scope, owned := header.VariablesScope()
	rows, found := export.VariablesFor(scope)
	if !owned || !found {
		return data
	}
	yamlnode.Patch(yamlnode.Root(document), []yamlnode.Field{model.VariablesField(rows)})
	encoded, err := yamlnode.Encode(document)
	if err != nil {
		return data
	}
	return encoded
}

func withSecretsBlock(data []byte, armored string) []byte {
	if armored == "" {
		return data
	}
	var block bytes.Buffer
	block.Write(data)
	if len(data) > 0 && !bytes.HasSuffix(data, []byte("\n")) {
		block.WriteByte('\n')
	}
	block.WriteString(secretsBlockHeader)
	for _, line := range strings.Split(strings.TrimRight(armored, "\n"), "\n") {
		block.WriteString(secretsBlockIndent + line + "\n")
	}
	return block.Bytes()
}

func isLocalOnlyDir(name string) bool {
	return name == localStateDir || name == gitDir || strings.HasPrefix(name, stagingDirPrefix)
}

func exportedAt() string {
	return timeutils.Now().UTC().Format(time.RFC3339)
}

func exportedBy() string {
	return exporterPrefix + properties.Properties().Application.Version
}
