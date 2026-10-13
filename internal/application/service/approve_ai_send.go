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
	scheduleAIApprovalTimeout usecase.ScheduleAIApprovalTimeout
	focusWindow               usecase.FocusWindow
	publishAIEvent            usecase.PublishAIEvent
}

func NewApproveAISendService(
	queueAIApproval usecase.QueueAIApproval,
	findAIApprovals usecase.FindAIApprovals,
	resolveAIApproval usecase.ResolveAIApproval,
	scheduleAIApprovalTimeout usecase.ScheduleAIApprovalTimeout,
	focusWindow usecase.FocusWindow,
	publishAIEvent usecase.PublishAIEvent,
) adapter.ApproveAISendService {
	return &approveAISend{
		queueAIApproval:           queueAIApproval,
		findAIApprovals:           findAIApprovals,
		resolveAIApproval:         resolveAIApproval,
		scheduleAIApprovalTimeout: scheduleAIApprovalTimeout,
		focusWindow:               focusWindow,
		publishAIEvent:            publishAIEvent,
	}
}

func (a *approveAISend) Request(ctx context.Context, request entity.AIApproval, sessionId string) (entity.AIApproval, error) {
	logger.Info(ctx, "Started", request, map[string]any{"sessionId": sessionId})

	approval := entity.NewAIApproval(idutils.NewUlid(), request.Client, request.Method, request.RawURL, request.Environment, timeutils.Now())
	resolved := a.queueAIApproval.Queue(ctx, approval, sessionId)
	a.scheduleAIApprovalTimeout.Schedule(ctx, approval.ID, entity.AIApprovalTimeout, func(ctx context.Context) {
		if err := a.resolve(ctx, approval.ID, entity.AIApproval.TimedOut); err != nil {
			logger.Error(ctx, err, "The approval timed out but its resolution could not be published", approval)
		}
	})
	a.focusWindow.Focus(ctx)
	if err := a.publishAIEvent.ApprovalRequested(ctx, approval); err != nil {
		return a.withdraw(ctx, approval.ID, resolved, err)
	}

	select {
	case approval = <-resolved:
	case <-ctx.Done():
		return a.withdraw(context.WithoutCancel(ctx), approval.ID, resolved, nil)
	}

	logger.Info(ctx, "Finished", approval, map[string]any{"sessionId": sessionId})
	return approval, nil
}

func (a *approveAISend) Answer(ctx context.Context, id string, approved bool) error {
	logger.Info(ctx, "Started", map[string]any{"id": id, "approved": approved})

	resolved, accepted := a.resolveAIApproval.Resolve(ctx, id, func(approval entity.AIApproval) (entity.AIApproval, bool) {
		return approval.Answered(approved)
	})
	if resolved == nil {
		return usecase.RejectUnknownApproval(id)
	}
	if !accepted {
		return usecase.RejectResolvedApproval(id)
	}
	if err := a.publishAIEvent.ApprovalResolved(ctx, *resolved); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", *resolved, map[string]any{"id": id, "approved": approved})
	return nil
}

func (a *approveAISend) Pending(ctx context.Context) []entity.AIApproval {
	logger.Info(ctx, "Started")

	pending := a.findAIApprovals.FindPending(ctx)

	logger.Info(ctx, "Finished", map[string]any{"pending": len(pending)})
	return pending
}

func (a *approveAISend) DenyAll(ctx context.Context) error {
	logger.Info(ctx, "Started")

	if err := a.denyEach(ctx, a.findAIApprovals.FindPending(ctx)); err != nil {
		return err
	}

	logger.Info(ctx, "Finished")
	return nil
}

func (a *approveAISend) DenySession(ctx context.Context, sessionId string) error {
	logger.Info(ctx, "Started", map[string]any{"sessionId": sessionId})

	if err := a.denyEach(ctx, a.findAIApprovals.FindSessionPending(ctx, sessionId)); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"sessionId": sessionId})
	return nil
}

func (a *approveAISend) withdraw(ctx context.Context, id string, resolved <-chan entity.AIApproval, cause error) (entity.AIApproval, error) {
	if err := a.resolve(ctx, id, entity.AIApproval.Denied); err != nil && cause == nil {
		cause = err
	}
	return <-resolved, cause
}

func (a *approveAISend) denyEach(ctx context.Context, approvals []entity.AIApproval) error {
	for _, approval := range approvals {
		if err := a.resolve(ctx, approval.ID, entity.AIApproval.Denied); err != nil {
			return err
		}
	}
	return nil
}

func (a *approveAISend) resolve(ctx context.Context, id string, resolve func(entity.AIApproval) (entity.AIApproval, bool)) error {
	resolved, accepted := a.resolveAIApproval.Resolve(ctx, id, resolve)
	if resolved == nil || !accepted {
		return nil
	}
	return a.publishAIEvent.ApprovalResolved(ctx, *resolved)
}
