package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type notifyAIChange struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	publishWorkspaceEvent usecase.PublishWorkspaceEvent
	publishVariableEvent  usecase.PublishVariableEvent
	publishAIEvent        usecase.PublishAIEvent
}

func NewNotifyAIChangeService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	publishWorkspaceEvent usecase.PublishWorkspaceEvent,
	publishVariableEvent usecase.PublishVariableEvent,
	publishAIEvent usecase.PublishAIEvent,
) adapter.NotifyAIChangeService {
	return &notifyAIChange{
		findWorkspaceRegistry: findWorkspaceRegistry,
		publishWorkspaceEvent: publishWorkspaceEvent,
		publishVariableEvent:  publishVariableEvent,
		publishAIEvent:        publishAIEvent,
	}
}

func (n *notifyAIChange) Notify(ctx context.Context, localId string, changes []entity.AIChange) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "changes": changes})

	for _, change := range changes {
		if err := n.publish(ctx, localId, change); err != nil {
			return err
		}
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "changes": changes})
	return nil
}

func (n *notifyAIChange) publish(ctx context.Context, localId string, change entity.AIChange) error {
	switch {
	case change.IsTree():
		return n.publishWorkspaceEvent.TreeChanged(ctx, localId)
	case change.IsRequest():
		return n.publishWorkspaceEvent.RequestChangedOnDisk(ctx, localId, change.RequestChange())
	case change.IsVariables():
		return n.publishVariableEvent.VariablesChanged(ctx, localId, []entity.ScopeRef{}, "")
	case change.IsHistory():
		return n.publishAIEvent.HistoryChanged(ctx, localId)
	case change.IsCookies():
		return n.publishWorkspaceEvent.CookiesChanged(ctx, localId)
	case change.IsWorkspaces():
		return n.publishWorkspacesChanged(ctx)
	default:
		return nil
	}
}

func (n *notifyAIChange) publishWorkspacesChanged(ctx context.Context) error {
	registry, err := n.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return err
	}
	return n.publishAIEvent.WorkspacesChanged(ctx, registry.ActiveLocalID)
}
