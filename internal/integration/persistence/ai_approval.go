package persistence

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const approvalIdAttribute = "id"

type queuedApproval struct {
	resolved  chan entity.AIApproval
	sessionId string
	approval  entity.AIApproval
}

type aiApprovalPersistence struct {
	approvals map[string]*queuedApproval
	order     []string
	mu        sync.Mutex
}

func NewAIApprovalPersistence() adapter.AIApprovalPersistence {
	return &aiApprovalPersistence{approvals: map[string]*queuedApproval{}}
}

func (a *aiApprovalPersistence) Queue(ctx context.Context, approval entity.AIApproval, sessionId string) <-chan entity.AIApproval {
	logger.Debug(ctx, "Started", approval, map[string]any{"sessionId": sessionId})

	queued := &queuedApproval{approval: approval, sessionId: sessionId, resolved: make(chan entity.AIApproval, 1)}
	a.mu.Lock()
	a.approvals[approval.ID] = queued
	a.order = append(a.order, approval.ID)
	a.mu.Unlock()

	logger.Debug(ctx, "Finished", approval, map[string]any{"sessionId": sessionId})
	return queued.resolved
}

func (a *aiApprovalPersistence) FindPending(ctx context.Context) []entity.AIApproval {
	logger.Debug(ctx, "Started")

	pending := a.pending(func(*queuedApproval) bool { return true })

	logger.Debug(ctx, "Finished", map[string]any{"pending": pending})
	return pending
}

func (a *aiApprovalPersistence) FindSessionPending(ctx context.Context, sessionId string) []entity.AIApproval {
	logger.Debug(ctx, "Started", map[string]any{"sessionId": sessionId})

	pending := a.pending(func(queued *queuedApproval) bool { return queued.sessionId == sessionId })

	logger.Debug(ctx, "Finished", map[string]any{"sessionId": sessionId, "pending": pending})
	return pending
}

func (a *aiApprovalPersistence) Resolve(ctx context.Context, id string, resolve func(entity.AIApproval) (entity.AIApproval, bool)) (*entity.AIApproval, error) {
	logger.Debug(ctx, "Started", map[string]any{"id": id})

	a.mu.Lock()
	defer a.mu.Unlock()
	queued, found := a.approvals[id]
	if !found {
		return nil, errs.BadRequestError("no pending send with this id", entity.AIInvalidArgumentsCode,
			errs.ErrorDetails{Attribute: approvalIdAttribute, Messages: []string{id}})
	}
	resolved, accepted := resolve(queued.approval)
	if !accepted {
		return nil, errs.ConflictError("this send was already approved, denied or timed out", "MCP-05409",
			errs.ErrorDetails{Attribute: approvalIdAttribute, Messages: []string{id}})
	}
	a.settle(id, queued, resolved)

	logger.Debug(ctx, "Finished", resolved, map[string]any{"id": id})
	return &resolved, nil
}

func (a *aiApprovalPersistence) Settle(ctx context.Context, id string, resolve func(entity.AIApproval) (entity.AIApproval, bool)) (*entity.AIApproval, bool) {
	logger.Debug(ctx, "Started", map[string]any{"id": id})

	a.mu.Lock()
	defer a.mu.Unlock()
	queued, found := a.approvals[id]
	if !found {
		logger.Debug(ctx, "Finished", map[string]any{"id": id, "settled": false})
		return nil, false
	}
	resolved, accepted := resolve(queued.approval)
	if !accepted {
		logger.Debug(ctx, "Finished", queued.approval, map[string]any{"id": id, "settled": false})
		return nil, false
	}
	a.settle(id, queued, resolved)

	logger.Debug(ctx, "Finished", resolved, map[string]any{"id": id, "settled": true})
	return &resolved, true
}

func (a *aiApprovalPersistence) settle(id string, queued *queuedApproval, resolved entity.AIApproval) {
	queued.approval = resolved
	a.order = a.withoutId(id)
	queued.resolved <- resolved
}

func (a *aiApprovalPersistence) pending(matches func(*queuedApproval) bool) []entity.AIApproval {
	a.mu.Lock()
	defer a.mu.Unlock()
	pending := make([]entity.AIApproval, 0, len(a.order))
	for _, id := range a.order {
		if queued := a.approvals[id]; matches(queued) {
			pending = append(pending, queued.approval)
		}
	}
	return pending
}

func (a *aiApprovalPersistence) withoutId(id string) []string {
	remaining := make([]string, 0, len(a.order))
	for _, candidate := range a.order {
		if candidate != id {
			remaining = append(remaining, candidate)
		}
	}
	return remaining
}
