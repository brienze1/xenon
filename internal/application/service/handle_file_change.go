package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/pkg/logger"
)

type handleFileChange struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	inspectFileChange     usecase.InspectFileChange
	watchWorkspace        usecase.WatchWorkspace
	publishWorkspaceEvent usecase.PublishWorkspaceEvent
}

func NewHandleFileChangeService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	inspectFileChange usecase.InspectFileChange,
	watchWorkspace usecase.WatchWorkspace,
	publishWorkspaceEvent usecase.PublishWorkspaceEvent,
) adapter.HandleFileChangeService {
	return &handleFileChange{
		findWorkspaceRegistry: findWorkspaceRegistry,
		inspectFileChange:     inspectFileChange,
		watchWorkspace:        watchWorkspace,
		publishWorkspaceEvent: publishWorkspaceEvent,
	}
}

func (h *handleFileChange) Handle(ctx context.Context, localId, path string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "path": path})

	workspace, err := h.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return err
	}
	if workspace.Missing {
		if err := h.watchWorkspace.Unwatch(ctx); err != nil {
			return err
		}
		return h.publishWorkspaceEvent.FolderMissing(ctx, localId)
	}
	if !workspace.Watches(path) {
		logger.Info(ctx, "Finished", map[string]any{"localId": localId, "path": path, "ignored": true})
		return nil
	}
	change, err := h.inspectFileChange.Inspect(ctx, workspace.Path, path)
	if err != nil {
		return err
	}
	if change.OwnWrite {
		logger.Info(ctx, "Finished", map[string]any{"localId": localId, "path": path, "ownWrite": true})
		return nil
	}
	for _, request := range change.Requests {
		if err := h.publishWorkspaceEvent.RequestChangedOnDisk(ctx, localId, request); err != nil {
			return err
		}
	}
	if err := h.publishWorkspaceEvent.TreeChanged(ctx, localId); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "path": path, "requests": len(change.Requests)})
	return nil
}
