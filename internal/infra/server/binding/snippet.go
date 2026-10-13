package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type SnippetService struct {
	snippet adapter.SnippetController
}

func NewSnippetService(snippet adapter.SnippetController) *SnippetService {
	return &SnippetService{snippet: snippet}
}

func (s *SnippetService) Languages(ctx context.Context) (dto.SnippetLanguageList, error) {
	return s.snippet.Languages(ctx)
}

func (s *SnippetService) Generate(ctx context.Context, request dto.GenerateSnippet) (*dto.Snippet, error) {
	return s.snippet.Generate(ctx, request)
}

func (s *SnippetService) Copy(ctx context.Context, request dto.GenerateSnippet) error {
	return s.snippet.Copy(ctx, request)
}
