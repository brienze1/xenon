package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ScriptToolsService interface {
	Check(ctx context.Context, text string) []entity.ScriptError
	Reference(ctx context.Context) ([]entity.ScriptMember, error)
	Snippets(ctx context.Context) ([]entity.ScriptSnippet, error)
}
