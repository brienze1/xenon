package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ApproveAISendService interface {
	Request(ctx context.Context, approval entity.AIApproval, sessionId string) (entity.AIApproval, error)
	Answer(ctx context.Context, id string, approved bool) error
	Deny(ctx context.Context, sessionId string) error
}
