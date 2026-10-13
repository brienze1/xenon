package webservice

import (
	"context"
	"time"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type aiApprovalTimer struct {
	timers adapter.Timers
}

func NewAIApprovalTimer(timers adapter.Timers) adapter.AIApprovalTimer {
	return &aiApprovalTimer{timers: timers}
}

func (a *aiApprovalTimer) Schedule(ctx context.Context, approvalId string, delay time.Duration, expire func(ctx context.Context)) {
	logger.Debug(ctx, "Started", map[string]any{"approvalId": approvalId, "delay": delay.String()})

	detached := context.WithoutCancel(ctx)
	a.timers.AfterFunc(delay, func() { expire(detached) })

	logger.Debug(ctx, "Finished", map[string]any{"approvalId": approvalId, "delay": delay.String()})
}
