package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type ScriptToolsController interface {
	Check(ctx context.Context, request dto.CheckScript) (dto.ScriptCheck, error)
	Reference(ctx context.Context) (dto.ScriptReference, error)
	Snippets(ctx context.Context) ([]dto.ScriptSnippet, error)
}
