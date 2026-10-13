package publisher

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

type variableEventPublisher struct {
	eventEmitter adapter.EventEmitter
}

func NewVariableEvent(eventEmitter adapter.EventEmitter) adapter.VariableEventPublisher {
	return &variableEventPublisher{eventEmitter: eventEmitter}
}

func (v *variableEventPublisher) VariablesChanged(ctx context.Context, localId string, scopes []entity.ScopeRef, reason string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "scopes": scopes, "reason": reason})

	v.eventEmitter.Emit(dto.VariablesChangedEvent, dto.NewVariablesChanged(localId, scopes, reason))

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "scopes": scopes, "reason": reason})
	return nil
}

func (v *variableEventPublisher) VaultStatusChanged(ctx context.Context, status entity.VaultStatus) error {
	logger.Debug(ctx, "Started", status)

	v.eventEmitter.Emit(dto.VaultStatusChangedEvent, dto.NewVaultStatusChanged(status))

	logger.Debug(ctx, "Finished", status)
	return nil
}
