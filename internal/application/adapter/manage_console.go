package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageConsoleService interface {
	List(ctx context.Context, query entity.ConsoleQuery) []entity.ConsoleEntry
	Get(ctx context.Context, id string) (*entity.ConsoleEntry, error)
	Clear(ctx context.Context)
}
