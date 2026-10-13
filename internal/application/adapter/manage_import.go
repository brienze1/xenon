package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageImportService interface {
	Detect(ctx context.Context, request entity.ImportRequest) (*entity.ImportSource, error)
	PickSources(ctx context.Context) ([]string, error)
	Clashes(ctx context.Context, plan entity.ImportPlan) ([]entity.Clash, error)
	Import(ctx context.Context, plan entity.ImportPlan) (*entity.ImportResult, error)
	Discard(ctx context.Context, sourceId string) error
}
