package webservice

import (
	"context"
	"time"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type aiApprovalTimerWebservice struct {
	timers adapter.Timers
}

func NewAIApprovalTimerWebservice(timers adapter.Timers) adapter.AIApprovalTimerWebservice {
	return &aiApprovalTimerWebservice{timers: timers}
}

func (a *aiApprovalTimerWebservice) Schedule(ctx context.Context, approvalId string, delay time.Duration, expire func(ctx context.Context) error) {
	logger.Debug(ctx, "Started", map[string]any{"approvalId": approvalId, "delay": delay.String()})

	detached := context.WithoutCancel(ctx)
	a.timers.AfterFunc(delay, func() { a.expire(detached, approvalId, expire) })

	logger.Debug(ctx, "Finished", map[string]any{"approvalId": approvalId, "delay": delay.String()})
}

func (a *aiApprovalTimerWebservice) expire(ctx context.Context, approvalId string, expire func(ctx context.Context) error) {
	if err := expire(ctx); err != nil {
		logger.Error(ctx, err, "The approval timed out but its resolution could not be published", map[string]any{"approvalId": approvalId})
	}
}
