package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveResponse interface {
	Save(ctx context.Context, tabId string, response entity.Response) error
}
