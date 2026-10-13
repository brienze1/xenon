package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type EnvironmentSetActiveController interface {
	Handle(ctx context.Context, request dto.SetActiveEnvironment) (*dto.ActiveEnvironment, error)
}
