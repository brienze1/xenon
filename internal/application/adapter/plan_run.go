package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PlanRunService interface {
	Plan(ctx context.Context, localId, itemId string) (*entity.RunPlan, error)
}
