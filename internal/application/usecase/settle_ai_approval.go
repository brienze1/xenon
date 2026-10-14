package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SettleAIApproval interface {
	// Settle reports false, without an error, when the approval is no longer pending or resolve refuses it.
	Settle(ctx context.Context, id string, resolve func(entity.AIApproval) (entity.AIApproval, bool)) (*entity.AIApproval, bool)
}
