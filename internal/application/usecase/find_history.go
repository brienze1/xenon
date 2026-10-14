package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindHistory interface {
	Find(ctx context.Context, localId string, caps entity.HistoryCaps) (*entity.History, error)
	FindEntry(ctx context.Context, localId, id string, caps entity.HistoryCaps) (*entity.HistoryEntry, error)
}
