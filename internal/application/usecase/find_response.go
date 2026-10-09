package usecase

import (
	"context"
	"io"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindResponse interface {
	Find(ctx context.Context, tabId string) (*entity.Response, error)
	OpenBody(ctx context.Context, tabId string) (io.ReadCloser, error)
}
