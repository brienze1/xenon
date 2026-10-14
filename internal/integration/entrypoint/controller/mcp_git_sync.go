package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

func loadMCPSyncStatus(ctx context.Context, manageSyncService appAdapter.ManageSyncService, localId string) (entity.SyncStatus, error) {
	request := dto.NewSyncWorkspaceReference(localId)
	err := validateSyncStatus(&request)
	if err != nil {
		return entity.SyncStatus{}, err
	}

	status, err := manageSyncService.Status(ctx, request.LocalId)
	if err != nil {
		return entity.SyncStatus{}, err
	}
	return *status, nil
}

func loadMCPSyncChanges(ctx context.Context, manageSyncService appAdapter.ManageSyncService, localId string) ([]entity.ChangeItem, error) {
	request := dto.NewSyncWorkspaceReference(localId)
	err := validateSyncChanges(&request)
	if err != nil {
		return nil, err
	}

	return manageSyncService.Changes(ctx, request.LocalId)
}

func loadMCPSyncSession(ctx context.Context, resolveConflictsService appAdapter.ResolveConflictsService, localId string) (*entity.SyncSession, error) {
	request := dto.NewSyncWorkspaceReference(localId)
	err := validateSyncSession(&request)
	if err != nil {
		return nil, err
	}

	return resolveConflictsService.Session(ctx, request.LocalId)
}
