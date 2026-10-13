package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type GenerateSnippet interface {
	Languages(ctx context.Context) []entity.SnippetLanguageSpec
	Generate(ctx context.Context, request entity.SnippetRequest, options entity.SnippetOptions) (*entity.Snippet, error)
}
