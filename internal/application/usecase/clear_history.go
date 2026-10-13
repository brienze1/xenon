package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/enums"
)

type ClearHistory interface {
	Clear(ctx context.Context, localId string, source enums.HistoryFilter) error
}
