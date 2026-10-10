package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type ConsoleController interface {
	List(ctx context.Context, request dto.ListConsole) ([]dto.ConsoleEntry, error)
	Clear(ctx context.Context) error
}
