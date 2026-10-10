package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SuggestVariablesService interface {
	Suggest(ctx context.Context, localId, requestId, prefix string) ([]entity.VariableSuggestion, error)
}
