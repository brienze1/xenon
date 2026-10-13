package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type RequestSaveController interface {
	Save(ctx context.Context, request dto.SaveRequest) (*dto.Request, error)
	SaveAs(ctx context.Context, request dto.SaveRequestAs) (*dto.Request, error)
}
