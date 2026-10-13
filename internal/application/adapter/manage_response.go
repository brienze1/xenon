package adapter

import (
	"context"
	"io"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageResponseService interface {
	LastResult(ctx context.Context, tabId string) (*entity.Response, error)
	OpenBody(ctx context.Context, tabId string) (*entity.Response, io.ReadCloser, error)
	SaveBody(ctx context.Context, tabId, path string) error
	Release(ctx context.Context, tabId string) error
	ReleaseAll(ctx context.Context) error
}
