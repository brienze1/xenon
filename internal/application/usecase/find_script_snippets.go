package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindScriptSnippets interface {
	FindSnippets(ctx context.Context) ([]entity.ScriptSnippet, error)
}
