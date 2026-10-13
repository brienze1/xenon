package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type RequestResponseController interface {
	LastResult(ctx context.Context, request dto.TabReference) (*dto.SendResult, error)
	ReleaseResponse(ctx context.Context, request dto.TabReference) error
	SaveResponseBody(ctx context.Context, request dto.SaveResponseBody) error
	OpenBody(ctx context.Context, request dto.TabReference) (*dto.ResponseBody, error)
	ReleaseAll(ctx context.Context) error
}
