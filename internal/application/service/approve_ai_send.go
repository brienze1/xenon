package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type approveAISend struct {
	queueAIApproval           usecase.QueueAIApproval
	findAIApprovals           usecase.FindAIApprovals
	resolveAIApproval         usecase.ResolveAIApproval
	settleAIApproval          usecase.SettleAIApproval
	scheduleAIApprovalTimeout usecase.ScheduleAIApprovalTimeout
	focusWindow               usecase.FocusWindow
	publishAIEvent            usecase.PublishAIEvent
}

func NewApproveAISendService(
	queueAIApproval usecase.QueueAIApproval,
	findAIApprovals usecase.FindAIApprovals,
	resolveAIApproval usecase.ResolveAIApproval,
	settleAIApproval usecase.SettleAIApproval,
	scheduleAIApprovalTimeout usecase.ScheduleAIApprovalTimeout,
	focusWindow usecase.FocusWindow,
	publishAIEvent usecase.PublishAIEvent,
) adapter.ApproveAISendService {
	return &approveAISend{
		queueAIApproval:           queueAIApproval,
		findAIApprovals:           findAIApprovals,
		resolveAIApproval:         resolveAIApproval,
		settleAIApproval:          settleAIApproval,
		scheduleAIApprovalTimeout: scheduleAIApprovalTimeout,
		focusWindow:               focusWindow,
		publishAIEvent:            publishAIEvent,
	}
}

func (a *approveAISend) Request(ctx context.Context, request entity.AIApproval, sessionId string) (entity.AIApproval, error) {
	logger.Info(ctx, "Started", request, map[string]any{"sessionId": sessionId})

	approval := request.Queued(idutils.NewUlid(), timeutils.Now())
	resolved := a.queueAIApproval.Queue(ctx, approval, sessionId)
	a.scheduleAIApprovalTimeout.Schedule(ctx, approval.ID, entity.AIApprovalTimeout, func(ctx context.Context) error {
		return a.settle(ctx, approval.ID, entity.AIApproval.TimedOut)
	})
	a.focusWindow.Focus(ctx)
	if err := a.publishAIEvent.ApprovalRequested(ctx, approval); err != nil {
		return a.withdraw(ctx, approval.ID, resolved, err)
	}

	answered, err := a.await(ctx, approval.ID, resolved)
	if err != nil {
		return answered, err
	}

	logger.Info(ctx, "Finished", request, answered, map[string]any{"sessionId": sessionId})
	return answered, nil
}

func (a *approveAISend) Answer(ctx context.Context, id string, approved bool) error {
	logger.Info(ctx, "Started", map[string]any{"id": id, "approved": approved})

	resolved, err := a.resolveAIApproval.Resolve(ctx, id, func(approval entity.AIApproval) (entity.AIApproval, bool) {
		return approval.Answered(approved)
	})
	if err != nil {
		return err
	}
	if err := a.publishAIEvent.ApprovalResolved(ctx, *resolved); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", *resolved, map[string]any{"id": id, "approved": approved})
	return nil
}

func (a *approveAISend) Deny(ctx context.Context, sessionId string) error {
	logger.Info(ctx, "Started", map[string]any{"sessionId": sessionId})

	if err := a.denyEach(ctx, a.pending(ctx, sessionId)); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"sessionId": sessionId})
	return nil
}

func (a *approveAISend) await(ctx context.Context, id string, resolved <-chan entity.AIApproval) (entity.AIApproval, error) {
	select {
	case approval := <-resolved:
		return approval, nil
	case <-ctx.Done():
		return a.withdraw(context.WithoutCancel(ctx), id, resolved, nil)
	}
}

func (a *approveAISend) withdraw(ctx context.Context, id string, resolved <-chan entity.AIApproval, cause error) (entity.AIApproval, error) {
	if err := a.settle(ctx, id, entity.AIApproval.Denied); err != nil {
		return <-resolved, err
	}
	return <-resolved, cause
}

func (a *approveAISend) pending(ctx context.Context, sessionId string) []entity.AIApproval {
	if entity.IsEverySession(sessionId) {
		return a.findAIApprovals.FindPending(ctx)
	}
	return a.findAIApprovals.FindSessionPending(ctx, sessionId)
}

func (a *approveAISend) denyEach(ctx context.Context, approvals []entity.AIApproval) error {
	for _, approval := range approvals {
		if err := a.settle(ctx, approval.ID, entity.AIApproval.Denied); err != nil {
			return err
		}
	}
	return nil
}

func (a *approveAISend) settle(ctx context.Context, id string, resolve func(entity.AIApproval) (entity.AIApproval, bool)) error {
	settled, ok := a.settleAIApproval.Settle(ctx, id, resolve)
	if !ok {
		return nil
	}
	return a.publishAIEvent.ApprovalResolved(ctx, *settled)
}
