package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const syncTimeLayout = "2006-01-02T15:04:05.000Z"

type SyncStatus struct {
	LastSyncAt   *string          `json:"last_sync_at,omitempty"`
	LastError    *SyncError       `json:"last_error,omitempty"`
	State        string           `json:"state"`
	Notice       string           `json:"notice,omitempty"`
	Settings     SyncLinkSettings `json:"settings"`
	LocalChanges int              `json:"local_changes"`
	Ahead        int              `json:"ahead"`
	Behind       int              `json:"behind"`
}

func NewSyncStatus(status entity.SyncStatus) SyncStatus {
	converted := SyncStatus{
		State:        status.State.String(),
		Notice:       status.Notice,
		Settings:     NewSyncLinkSettings(status.Settings),
		LocalChanges: status.LocalChanges,
		Ahead:        status.Ahead,
		Behind:       status.Behind,
	}
	if status.LastSyncAt != nil {
		at := status.LastSyncAt.UTC().Format(syncTimeLayout)
		converted.LastSyncAt = &at
	}
	if status.LastError != nil {
		failure := NewSyncError(*status.LastError)
		converted.LastError = &failure
	}
	return converted
}
