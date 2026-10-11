package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type ConsoleService struct {
	console adapter.ConsoleController
}

func NewConsoleService(console adapter.ConsoleController) *ConsoleService {
	return &ConsoleService{console: console}
}

func (c *ConsoleService) List(ctx context.Context, request dto.ListConsole) ([]dto.ConsoleEntry, error) {
	return c.console.List(ctx, request)
}

func (c *ConsoleService) Get(ctx context.Context, request dto.ConsoleEntryReference) (dto.ConsoleEntry, error) {
	return c.console.Get(ctx, request)
}

func (c *ConsoleService) Clear(ctx context.Context) error {
	return c.console.Clear(ctx)
}
