package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateSyncStatus(request *dto.SyncWorkspaceReference) error {
	return validator.CustomValidator().Struct("GIT-67400", request)
}

func validateSyncChanges(request *dto.SyncWorkspaceReference) error {
	return validator.CustomValidator().Struct("GIT-68400", request)
}

func validateSaveSyncSettings(request *dto.SaveSyncSettings) error {
	return validator.CustomValidator().Struct("GIT-71400", request)
}

func validateSyncSession(request *dto.SyncWorkspaceReference) error {
	return validator.CustomValidator().Struct("GIT-33400", request)
}

func validateResolveSyncConflict(request *dto.ResolveSyncConflict) error {
	return validator.CustomValidator().Struct("GIT-29400", request)
}

func validateFinishPull(request *dto.SyncSessionReference) error {
	return validator.CustomValidator().Struct("GIT-31400", request)
}

func validateCancelPull(request *dto.SyncSessionReference) error {
	return validator.CustomValidator().Struct("GIT-32400", request)
}

func validateCommitSyncChanges(request *dto.CommitSyncChanges) error {
	return validator.CustomValidator().Struct("GIT-00400", request)
}

func validatePullWorkspace(request *dto.SyncWorkspaceReference) error {
	return validator.CustomValidator().Struct("GIT-69400", request)
}

func validatePushWorkspace(request *dto.SyncWorkspaceReference) error {
	return validator.CustomValidator().Struct("GIT-70400", request)
}

func validateLinkRepository(request *dto.LinkRepository) error {
	return validator.CustomValidator().Struct("GIT-63400", request)
}

func validateCloneWorkspace(request *dto.CloneWorkspace) error {
	return validator.CustomValidator().Struct("GIT-39400", request)
}

func validateUnlinkRepository(request *dto.UnlinkRepository) error {
	return validator.CustomValidator().Struct("GIT-40400", request)
}
