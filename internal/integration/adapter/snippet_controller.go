package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type SnippetController interface {
	Languages(ctx context.Context) (dto.SnippetLanguageList, error)
	Generate(ctx context.Context, request dto.GenerateSnippet) (*dto.Snippet, error)
	Copy(ctx context.Context, request dto.GenerateSnippet) error
}
