package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type CheckScript interface {
	Check(ctx context.Context, text string) []entity.ScriptError
}
