package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type InspectVariablesService interface {
	Inspect(ctx context.Context, localId string, request entity.Request) (*entity.VariableInspection, error)
	Suggest(ctx context.Context, localId, requestId, prefix string) ([]entity.VariableSuggestion, error)
}
