package persistence

import (
	"context"
	"net/http"
	"os"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/internal/integration/persistence/yamlnode"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

type requestPersistence struct {
	index *WorkspaceIndex
}

func NewRequestPersistence(index *WorkspaceIndex) adapter.RequestPersistence {
	return &requestPersistence{index: index}
}

func (r *requestPersistence) Find(ctx context.Context, root, id string) (*entity.Request, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id})
	r.index.mu.Lock()
	defer r.index.mu.Unlock()

	path, err := r.path(root, id)
	if err != nil {
		return nil, err
	}
	request, err := readRequestEntity(id, path)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", request)
	return request, nil
}

func (r *requestPersistence) Save(ctx context.Context, root string, request entity.Request, baseRevision string, force bool) (*entity.Request, error) {
	logger.Debug(ctx, "Started", request, map[string]any{"root": root, "baseRevision": baseRevision, "force": force})
	r.index.mu.Lock()
	defer r.index.mu.Unlock()

	path, err := r.path(root, request.ID)
	if err != nil {
		return nil, err
	}
	current, err := os.ReadFile(path)
	if err != nil {
		return nil, errs.FileError(err, "REQ-12500")
	}
	if !force && revisionOf(current) != baseRevision {
		return nil, errs.ClientError("the request changed on disk since it was opened", "REQ-02409", http.StatusConflict)
	}
	document, err := yamlnode.Parse(current)
	if err != nil {
		document = yamlnode.NewDocument()
	}
	stored := model.NewRequest(request)
	yamlnode.Patch(yamlnode.Root(document), stored.Fields())
	data, err := yamlnode.Encode(document)
	if err != nil {
		return nil, errs.InternalServerError(err, "REQ-14500", "the request could not be encoded")
	}
	if err := fileutils.WriteFileAtomic(path, data, workspaceFileMode); err != nil {
		return nil, errs.FileError(err, "REQ-13500")
	}
	r.index.recordWrite(path, data)
	saved, err := readRequestEntity(request.ID, path)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", saved)
	return saved, nil
}

func (r *requestPersistence) path(root, id string) (string, error) {
	_, item, found, err := r.index.resolve(root, id)
	if err != nil {
		return "", err
	}
	if !found || item.kind != itemkind.Request {
		return "", errs.NotFoundError("the request does not exist", "REQ-01404")
	}
	return item.path, nil
}
