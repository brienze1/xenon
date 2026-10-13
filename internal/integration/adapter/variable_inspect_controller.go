package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type VariableInspectController interface {
	Inspect(ctx context.Context, request dto.InspectVariables) (*dto.VariableInspection, error)
	Suggest(ctx context.Context, request dto.SuggestVariables) (dto.VariableSuggestionList, error)
}
