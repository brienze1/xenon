package persistence

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

func (t *treePersistence) ApplySyncedFiles(ctx context.Context, root string, writes map[string][]byte, removals []string) ([]entity.RequestChange, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "writes": len(writes), "removals": len(removals)})
	t.index.Lock(ctx)
	defer t.index.Unlock(ctx)

	before, err := t.index.Current(ctx, root)
	if err != nil {
		return nil, err
	}
	affected := map[string]bool{}
	for path, data := range writes {
		target := filepath.Join(root, filepath.FromSlash(path))
		markRequests(affected, before.RequestIDsInside(target))
		if err := os.MkdirAll(filepath.Dir(target), workspaceDirMode); err != nil {
			return nil, errs.FileError(err, "GIT-76500")
		}
		if err := fileutils.WriteFileAtomic(target, data, workspaceFileMode); err != nil {
			return nil, errs.FileError(err, "GIT-77500")
		}
		t.index.RecordWrite(ctx, target, data)
	}
	for _, path := range removals {
		target := filepath.Join(root, filepath.FromSlash(path))
		markRequests(affected, before.RequestIDsInside(target))
		t.index.RecordMove(ctx, target, target)
		if err := os.Remove(target); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, errs.FileError(err, "GIT-78500")
		}
		removeEmptyParents(root, filepath.Dir(target))
	}
	after, err := t.index.Refresh(ctx, root)
	if err != nil {
		return nil, err
	}
	for path := range writes {
		markRequests(affected, after.RequestIDsInside(filepath.Join(root, filepath.FromSlash(path))))
	}
	changes := make([]entity.RequestChange, 0, len(affected))
	for _, id := range sortedKeys(affected) {
		item, _ := after.Item(id)
		changes = append(changes, requestChange(id, item))
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "requests": len(changes)})
	return changes, nil
}

func (t *treePersistence) IdentifyItems(ctx context.Context, files map[string][]byte) map[string]entity.ItemRef {
	logger.Debug(ctx, "Started", map[string]any{"files": len(files)})

	refs := map[string]entity.ItemRef{}
	for path, content := range files {
		var header model.ItemHeader
		if err := yaml.Unmarshal(content, &header); err == nil && header.IsIdentified() {
			refs[path] = header.ToEntity()
		}
	}

	logger.Debug(ctx, "Finished", map[string]any{"identified": len(refs)})
	return refs
}

func (t *treePersistence) DecodeDocument(ctx context.Context, content []byte) (*entity.DocumentNode, error) {
	logger.Debug(ctx, "Started", map[string]any{"bytes": len(content)})

	document, err := yamlnode.Parse(content)
	if err != nil {
		return nil, errs.UnprocessableEntityError("the file cannot be merged because it is not valid YAML: "+err.Error(), "GIT-79422")
	}
	node := model.NewDocumentNode(document)

	logger.Debug(ctx, "Finished", map[string]any{"bytes": len(content)})
	return node, nil
}

func (t *treePersistence) EncodeDocument(ctx context.Context, node *entity.DocumentNode) ([]byte, error) {
	logger.Debug(ctx, "Started")

	data, err := yamlnode.Encode(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{model.DocumentYaml(node)}})
	if err != nil {
		return nil, errs.InternalServerError(err, "GIT-80500", "the merged file could not be encoded")
	}

	logger.Debug(ctx, "Finished", map[string]any{"bytes": len(data)})
	return data, nil
}

func markRequests(affected map[string]bool, ids []string) {
	for _, id := range ids {
		affected[id] = true
	}
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func removeEmptyParents(root, dir string) {
	for filepath.Dir(dir) != root && strings.HasPrefix(dir, root+string(filepath.Separator)) {
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}
