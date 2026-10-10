package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type RunScript interface {
	Run(ctx context.Context, script entity.ScriptRef, settings entity.AppSettings) entity.ScriptRun
}
