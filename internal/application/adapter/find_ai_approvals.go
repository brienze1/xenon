package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindAIApprovalsService interface {
	FindPending(ctx context.Context) []entity.AIApproval
}
