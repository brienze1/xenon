package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageConsoleService interface {
	List(ctx context.Context, afterId string, limit int) []entity.ConsoleEntry
	Clear(ctx context.Context)
}
