package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ApproveAISendService interface {
	Request(ctx context.Context, approval entity.AIApproval, sessionId string) (entity.AIApproval, error)
	Answer(ctx context.Context, id string, approved bool) error
	Pending(ctx context.Context) []entity.AIApproval
	DenyAll(ctx context.Context) error
	DenySession(ctx context.Context, sessionId string) error
}
