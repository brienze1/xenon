package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const GitStatusChangedEvent = "git:status-changed"

type GitStatusChanged struct {
	LocalId string     `json:"local_id"`
	Status  SyncStatus `json:"status"`
}

func NewGitStatusChanged(localId string, status entity.SyncStatus) GitStatusChanged {
	return GitStatusChanged{LocalId: localId, Status: NewSyncStatus(status)}
}
