package persistence

import (
	"context"
	"os"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type requestPersistence struct {
	index adapter.WorkspaceIndex
}

func NewRequestPersistence(index adapter.WorkspaceIndex) adapter.RequestPersistence {
	return &requestPersistence{index: index}
}

func (r *requestPersistence) Find(ctx context.Context, root, id string) (*entity.Request, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id})
	r.index.Lock()
	defer r.index.Unlock()

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

func (r *requestPersistence) FindRevision(ctx context.Context, root, id string) (*entity.RequestChange, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": root, "id": id})
	r.index.Lock()
	defer r.index.Unlock()

	path, err := r.path(root, id)
	if err != nil {
		return nil, err
	}
	current, err := os.ReadFile(path)
	if err != nil {
		return nil, errs.FileError(err, "REQ-12500")
	}
	revision := entity.RequestChange{RequestID: id, Revision: revisionOf(current)}

	logger.Debug(ctx, "Finished", map[string]any{"root": root, "id": id, "revision": revision.Revision})
	return &revision, nil
}

func (r *requestPersistence) Save(ctx context.Context, root string, request entity.Request) (*entity.Request, error) {
	logger.Debug(ctx, "Started", request, map[string]any{"root": root})
	r.index.Lock()
	defer r.index.Unlock()

	path, err := r.path(root, request.ID)
	if err != nil {
		return nil, err
	}
	current, err := os.ReadFile(path)
	if err != nil {
		return nil, errs.FileError(err, "REQ-17500")
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
	r.index.RecordWrite(path, data)
	saved, err := readRequestEntity(request.ID, path)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", saved)
	return saved, nil
}

func (r *requestPersistence) path(root, id string) (string, error) {
	_, item, found, err := r.index.Resolve(root, id)
	if err != nil {
		return "", err
	}
	if !found || item.Kind != itemkind.Request {
		return "", errs.NotFoundError("the request does not exist", "REQ-01404")
	}
	return item.Path, nil
}

func readRequestEntity(id, path string) (*entity.Request, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errs.FileError(err, "REQ-16500")
	}
	stored, reason := readRequest(path)
	if reason != "" {
		return nil, errs.UnprocessableEntityError(reason, "REQ-15422")
	}
	request := stored.ToStoredEntity(id, revisionOf(data))
	return &request, nil
}
