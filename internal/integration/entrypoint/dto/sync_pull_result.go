package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SyncPullResult struct {
	Session *SyncSession `json:"session,omitempty"`
	Status  SyncStatus   `json:"status"`
}

func NewSyncPullResult(result entity.PullResult) SyncPullResult {
	converted := SyncPullResult{Status: NewSyncStatus(result.Status)}
	if result.Session != nil {
		session := NewSyncSession(*result.Session)
		converted.Session = &session
	}
	return converted
}
