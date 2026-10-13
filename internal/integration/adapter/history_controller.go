package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type HistoryController interface {
	List(ctx context.Context, request dto.ListHistory) (dto.HistoryEntryList, error)
	Get(ctx context.Context, request dto.HistoryReference) (*dto.HistoryEntryDetail, error)
	Clear(ctx context.Context, request dto.ClearHistory) error
}
