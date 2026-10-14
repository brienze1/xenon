package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

type ManageHistoryService interface {
	List(ctx context.Context, localId string, query entity.HistoryQuery) ([]entity.HistoryEntry, error)
	Get(ctx context.Context, localId, id string) (*entity.HistoryEntry, error)
	Clear(ctx context.Context, localId string, source enums.HistoryFilter) error
}
