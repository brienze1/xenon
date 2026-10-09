package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveHistoryEntry interface {
	Save(ctx context.Context, localId string, entry entity.HistoryEntry) error
}
