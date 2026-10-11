package persistence

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/logger"
)

func (t *treePersistence) Inspect(ctx context.Context, root, path string) (*entity.FileChange, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "path": path})
	t.index.Lock(ctx)
	defer t.index.Unlock(ctx)

	if t.isOwnChange(ctx, path) {
		logger.Debug(ctx, "Finished", map[string]any{"root": root, "path": path, "ownWrite": true})
		return &entity.FileChange{OwnWrite: true}, nil
	}
	affected := t.requestsAt(ctx, root, path)
	index, err := t.index.Refresh(ctx, root)
	if err != nil {
		return nil, err
	}
	change := variableChangesAt(root, path)
	change.Requests = make([]entity.RequestChange, 0, len(affected))
	for _, id := range affected {
		item, _ := index.Item(id)
		change.Requests = append(change.Requests, requestChange(id, item))
	}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "path": path, "requests": len(change.Requests), "variableScopes": len(change.VariableScopes)})
	return &change, nil
}

func (t *treePersistence) isOwnChange(ctx context.Context, path string) bool {
	if t.index.IsOwnMove(ctx, path) {
		return true
	}
	data, err := os.ReadFile(path)
	return err == nil && t.index.IsOwnWrite(ctx, path, data)
}

func (t *treePersistence) requestsAt(ctx context.Context, root, path string) []string {
	index, found := t.index.Cached(ctx, root)
	if !found || filepath.Clean(path) == filepath.Clean(root) {
		return []string{}
	}
	return index.RequestIDsInside(path)
}

func requestChange(id string, item entity.IndexedItem) entity.RequestChange {
	data, err := os.ReadFile(item.Path)
	if err != nil {
		return entity.RequestChange{RequestID: id, Deleted: true}
	}
	return entity.RequestChange{RequestID: id, Revision: revisionOf(data)}
}

func variableChangesAt(root, path string) entity.FileChange {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return entity.FileChange{}
	}
	segments := strings.Split(filepath.ToSlash(relative), "/")
	switch {
	case relative == workspaceFile:
		return entity.FileChange{VariableScopes: []entity.ScopeRef{entity.NewGlobalsScopeRef()}}
	case len(segments) == 3 && segments[0] == collectionsDir && segments[2] == collectionMarker:
		return entity.FileChange{VariableScopes: collectionScopeAt(path)}
	case segments[0] == environmentsDir:
		return entity.FileChange{EnvironmentsChanged: true, VariableScopes: environmentScopeAt(path)}
	default:
		return entity.FileChange{}
	}
}

func collectionScopeAt(path string) []entity.ScopeRef {
	stored, reason := readCollection(path, itemkind.Collection)
	if reason != "" || strings.TrimSpace(stored.Id) == "" {
		return []entity.ScopeRef{}
	}
	return []entity.ScopeRef{entity.NewCollectionScopeRef(stored.Id)}
}

func environmentScopeAt(path string) []entity.ScopeRef {
	if !isRegularFile(path) {
		return []entity.ScopeRef{}
	}
	environment := readEnvironment(filepath.Dir(path), filepath.Base(path))
	if environment.IsMalformed() {
		return []entity.ScopeRef{}
	}
	return []entity.ScopeRef{environment.ScopeRef()}
}
