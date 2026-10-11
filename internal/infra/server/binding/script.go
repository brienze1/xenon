package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type ScriptService struct {
	scriptTools adapter.ScriptToolsController
}

func NewScriptService(scriptTools adapter.ScriptToolsController) *ScriptService {
	return &ScriptService{scriptTools: scriptTools}
}

func (s *ScriptService) Check(ctx context.Context, request dto.CheckScript) (dto.ScriptCheck, error) {
	return s.scriptTools.Check(ctx, request)
}

func (s *ScriptService) Reference(ctx context.Context) (dto.ScriptReference, error) {
	return s.scriptTools.Reference(ctx)
}

func (s *ScriptService) Snippets(ctx context.Context) ([]dto.ScriptSnippet, error) {
	return s.scriptTools.Snippets(ctx)
}
