package webservice

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/snippets"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type snippetWebservice struct{}

func NewSnippetWebservice() adapter.SnippetWebservice {
	return &snippetWebservice{}
}

func (s *snippetWebservice) Languages(ctx context.Context) []entity.SnippetLanguageSpec {
	logger.Debug(ctx, "Started", nil)

	languages := snippets.Languages()

	logger.Debug(ctx, "Finished", map[string]any{"result": len(languages)})
	return languages
}

func (s *snippetWebservice) Generate(ctx context.Context, request entity.SnippetRequest, options entity.SnippetOptions) (*entity.Snippet, error) {
	logger.Debug(ctx, "Started", map[string]any{"request": request, "language": options.Language})

	text, found := snippets.Generate(request, options)
	if !found {
		return nil, errs.BadRequestError("Snippets are not available for the language "+options.Language.String(), "SNP-02400")
	}
	snippet := entity.Snippet{Text: text, Notices: request.Notices}

	logger.Debug(ctx, "Finished", map[string]any{"request": request, "language": options.Language, "result": snippet})
	return &snippet, nil
}
