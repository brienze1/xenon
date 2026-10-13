package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type RequestGetController interface {
	Get(ctx context.Context, request dto.OpenRequest) (*dto.Request, error)
}
