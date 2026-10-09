package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type HistoryService struct {
	history adapter.HistoryController
}

func NewHistoryService(history adapter.HistoryController) *HistoryService {
	return &HistoryService{history: history}
}

func (h *HistoryService) List(ctx context.Context, request dto.ListHistory) (dto.HistoryEntryList, error) {
	return h.history.List(ctx, request)
}

func (h *HistoryService) Get(ctx context.Context, request dto.HistoryReference) (*dto.HistoryEntryDetail, error) {
	return h.history.Get(ctx, request)
}

func (h *HistoryService) Clear(ctx context.Context, request dto.ClearHistory) error {
	return h.history.Clear(ctx, request)
}
