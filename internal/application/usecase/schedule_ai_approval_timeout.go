package usecase

import (
	"context"
	"time"
)

type ScheduleAIApprovalTimeout interface {
	// Schedule runs expire after delay, detached from ctx's cancellation.
	Schedule(ctx context.Context, approvalId string, delay time.Duration, expire func(ctx context.Context))
}
