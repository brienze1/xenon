package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindScriptReference interface {
	FindReference(ctx context.Context) ([]entity.ScriptMember, error)
}
