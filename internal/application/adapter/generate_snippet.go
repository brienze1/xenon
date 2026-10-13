package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type GenerateSnippetService interface {
	Languages(ctx context.Context) []entity.SnippetLanguageSpec
	Generate(ctx context.Context, input entity.SnippetInput) (*entity.Snippet, error)
	Copy(ctx context.Context, input entity.SnippetInput) error
}
