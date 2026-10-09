package persistence

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/internal/integration/persistence/yamlnode"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	workspaceFile       = "workspace.yaml"
	collectionsDir      = "collections"
	environmentsDir     = "environments"
	gitignoreFile       = ".gitignore"
	gitignoreHeader     = "# Xenon — local-only files (never commit)"
	gitignoreBlock      = gitignoreHeader + "\n.xenon/\n*.xenon-tmp\n"
	workspaceDirMode    = 0o755
	workspaceFileMode   = 0o644
	latestFormatVersion = 1
)

type workspacePersistence struct {
	index *WorkspaceIndex
}

func NewWorkspacePersistence(index *WorkspaceIndex) adapter.WorkspacePersistence {
	return &workspacePersistence{index: index}
}

func (w *workspacePersistence) ValidateNew(ctx context.Context, root string, registry entity.AppRegistry) error {
	logger.Debug(ctx, "Started", map[string]any{"root": root})

	if registry.Overlaps(root) {
		return errs.ClientError("the folder is, contains or is inside a registered workspace", "WKS-02409", http.StatusConflict)
	}
	info, err := os.Stat(root)
	if errors.Is(err, fs.ErrNotExist) {
		logger.Debug(ctx, "Finished", map[string]any{"root": root})
		return nil
	}
	if err != nil {
		return errs.FileError(err, "WKS-09500")
	}
	entries, _ := os.ReadDir(root)
	if !info.IsDir() || len(entries) > 0 {
		return errs.ClientError("the folder is not empty", "WKS-01409", http.StatusConflict)
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root})
	return nil
}

func (w *workspacePersistence) ValidateExisting(ctx context.Context, root, localId string, registry entity.AppRegistry) error {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "localId": localId})

	if registry.OverlapsOther(root, localId) {
		return errs.ClientError("the folder is, contains or is inside another registered workspace", "WKS-02409", http.StatusConflict)
	}
	if !isDirectory(root) {
		return errs.NotFoundError("the folder does not exist", "WKS-03404")
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "localId": localId})
	return nil
}

func (w *workspacePersistence) Save(ctx context.Context, root string, workspace entity.Workspace) error {
	logger.Debug(ctx, "Started", workspace, map[string]any{"root": root})

	document, _, err := readWorkspace(root)
	if err != nil {
		return err
	}
	stored := model.NewWorkspace(workspace)
	yamlnode.Patch(yamlnode.Root(document), stored.Fields())
	data, err := yamlnode.Encode(document)
	if err != nil {
		return errs.InternalServerError(err, "WKS-25500", "workspace.yaml could not be encoded")
	}
	path := filepath.Join(root, workspaceFile)
	if err := fileutils.WriteFileAtomic(path, data, workspaceFileMode); err != nil {
		return errs.FileError(err, "WKS-26500")
	}
	w.index.recordWrite(path, data)

	logger.Debug(ctx, "Finished", workspace, map[string]any{"root": root})
	return nil
}

func (w *workspacePersistence) Create(ctx context.Context, root string, workspace entity.Workspace) error {
	logger.Debug(ctx, "Started", workspace, map[string]any{"root": root})

	for _, dir := range []string{collectionsDir, environmentsDir} {
		if err := os.MkdirAll(filepath.Join(root, dir), workspaceDirMode); err != nil {
			return errs.FileError(err, "WKS-10500")
		}
	}
	document := yamlnode.NewDocument()
	stored := model.NewWorkspace(workspace)
	yamlnode.Patch(yamlnode.Root(document), stored.Fields())
	data, err := yamlnode.Encode(document)
	if err != nil {
		return errs.InternalServerError(err, "WKS-20500", "workspace.yaml could not be encoded")
	}
	if err := fileutils.WriteFileAtomic(filepath.Join(root, workspaceFile), data, workspaceFileMode); err != nil {
		return errs.FileError(err, "WKS-17500")
	}
	if err := ensureGitignore(root); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", workspace, map[string]any{"root": root})
	return nil
}

func (w *workspacePersistence) Load(ctx context.Context, root string) (*entity.Workspace, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root})

	_, stored, err := readWorkspace(root)
	if err != nil {
		return nil, err
	}
	workspace := stored.ToEntity()

	logger.Debug(ctx, "Finished", workspace)
	return &workspace, nil
}

func readWorkspace(root string) (*yaml.Node, *model.Workspace, error) {
	if !isDirectory(root) {
		return nil, nil, errs.NotFoundError("the workspace folder is missing", "WKS-06404")
	}
	data, err := os.ReadFile(filepath.Join(root, workspaceFile))
	if err != nil {
		return nil, nil, errs.FileError(err, "WKS-03404")
	}
	document, err := yamlnode.Parse(data)
	if err != nil {
		return nil, nil, errs.ClientError("workspace.yaml cannot be parsed: "+err.Error(), "WKS-15422", http.StatusUnprocessableEntity)
	}
	var stored model.Workspace
	if err := document.Decode(&stored); err != nil {
		return nil, nil, errs.ClientError("workspace.yaml has invalid fields: "+err.Error(), "WKS-16422", http.StatusUnprocessableEntity)
	}
	if stored.Xenon > latestFormatVersion {
		return nil, nil, errs.ClientError("workspace.yaml was written by a newer Xenon", "WKS-04422", http.StatusUnprocessableEntity)
	}
	return document, &stored, nil
}

func ensureGitignore(root string) error {
	path := filepath.Join(root, gitignoreFile)
	current, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return errs.FileError(err, "WKS-18500")
	}
	if strings.Contains(string(current), gitignoreHeader) {
		return nil
	}
	content := string(current)
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if err := fileutils.WriteFileAtomic(path, []byte(content+gitignoreBlock), workspaceFileMode); err != nil {
		return errs.FileError(err, "WKS-19500")
	}
	return nil
}
