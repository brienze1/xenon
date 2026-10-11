package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type scriptTools struct {
	checkScript         usecase.CheckScript
	findScriptReference usecase.FindScriptReference
	findScriptSnippets  usecase.FindScriptSnippets
}

func NewScriptToolsService(checkScript usecase.CheckScript, findScriptReference usecase.FindScriptReference, findScriptSnippets usecase.FindScriptSnippets) adapter.ScriptToolsService {
	return &scriptTools{checkScript: checkScript, findScriptReference: findScriptReference, findScriptSnippets: findScriptSnippets}
}

func (s *scriptTools) Check(ctx context.Context, text string) []entity.ScriptError {
	logger.Info(ctx, "Started", map[string]any{"length": len(text)})

	errors := s.checkScript.Check(ctx, text)

	logger.Info(ctx, "Finished", map[string]any{"length": len(text), "errors": len(errors)})
	return errors
}

func (s *scriptTools) Reference(ctx context.Context) ([]entity.ScriptMember, error) {
	logger.Info(ctx, "Started")

	members, err := s.findScriptReference.FindReference(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"members": len(members)})
	return members, nil
}

func (s *scriptTools) Snippets(ctx context.Context) ([]entity.ScriptSnippet, error) {
	logger.Info(ctx, "Started")

	snippets, err := s.findScriptSnippets.FindSnippets(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"snippets": len(snippets)})
	return snippets, nil
}
