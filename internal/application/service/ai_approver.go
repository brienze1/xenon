package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type aiApprover struct {
	queueAIApproval           usecase.QueueAIApproval
	settleAIApproval          usecase.SettleAIApproval
	scheduleAIApprovalTimeout usecase.ScheduleAIApprovalTimeout
	focusWindow               usecase.FocusWindow
	publishAIEvent            usecase.PublishAIEvent
}

func newAIApprover(
	queueAIApproval usecase.QueueAIApproval,
	settleAIApproval usecase.SettleAIApproval,
	scheduleAIApprovalTimeout usecase.ScheduleAIApprovalTimeout,
	focusWindow usecase.FocusWindow,
	publishAIEvent usecase.PublishAIEvent,
) *aiApprover {
	return &aiApprover{
		queueAIApproval:           queueAIApproval,
		settleAIApproval:          settleAIApproval,
		scheduleAIApprovalTimeout: scheduleAIApprovalTimeout,
		focusWindow:               focusWindow,
		publishAIEvent:            publishAIEvent,
	}
}

func (a *aiApprover) request(ctx context.Context, request entity.AIApproval, sessionId string) (entity.AIApproval, error) {
	approval := request.Queued(idutils.NewUlid(), timeutils.Now())
	resolved := a.queueAIApproval.Queue(ctx, approval, sessionId)
	a.scheduleAIApprovalTimeout.Schedule(ctx, approval.ID, entity.AIApprovalTimeout, func(ctx context.Context) error {
		return a.settle(ctx, approval.ID, entity.AIApproval.TimedOut)
	})
	a.focusWindow.Focus(ctx)
	if err := a.publishAIEvent.ApprovalRequested(ctx, approval); err != nil {
		return a.withdraw(ctx, approval.ID, resolved, err)
	}
	return a.await(ctx, approval.ID, resolved)
}

func (a *aiApprover) await(ctx context.Context, id string, resolved <-chan entity.AIApproval) (entity.AIApproval, error) {
	select {
	case approval := <-resolved:
		return approval, nil
	case <-ctx.Done():
		return a.withdraw(context.WithoutCancel(ctx), id, resolved, nil)
	}
}

func (a *aiApprover) withdraw(ctx context.Context, id string, resolved <-chan entity.AIApproval, cause error) (entity.AIApproval, error) {
	if err := a.settle(ctx, id, entity.AIApproval.Denied); err != nil {
		return <-resolved, err
	}
	return <-resolved, cause
}

func (a *aiApprover) settle(ctx context.Context, id string, resolve func(entity.AIApproval) (entity.AIApproval, bool)) error {
	settled, ok := a.settleAIApproval.Settle(ctx, id, resolve)
	if !ok {
		return nil
	}
	return a.publishAIEvent.ApprovalResolved(ctx, *settled)
}
