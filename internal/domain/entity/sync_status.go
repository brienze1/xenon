package entity

import (
	"fmt"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/syncstate"
)

const sharedSettingsNotice = "The shared sync settings changed to %s (branch %s, subfolder %q); re-link to use them - this clone keeps its settings"

type SyncStatus struct {
	LastSyncAt   *time.Time
	LastError    *SyncError
	State        enums.SyncState
	Notice       string
	Settings     LocalSyncSettings
	LocalChanges int
	Ahead        int
	Behind       int
}

func NotLinkedSyncStatus() SyncStatus {
	return SyncStatus{State: syncstate.NotLinked}
}

func SyncingSyncStatus(settings LocalSyncSettings) SyncStatus {
	return SyncStatus{State: syncstate.Syncing, Settings: settings}
}

func NewSyncStatus(settings LocalSyncSettings, heads BranchHeads, localChanges int) SyncStatus {
	status := SyncStatus{Settings: settings, LocalChanges: localChanges, Ahead: heads.Ahead, Behind: heads.Behind}
	switch {
	case heads.Ahead > 0 && heads.Behind > 0:
		status.State = syncstate.Diverged
	case heads.Behind > 0:
		status.State = syncstate.Behind
	case heads.Ahead > 0:
		status.State = syncstate.Ahead
	case localChanges > 0:
		status.State = syncstate.LocalChanges
	default:
		status.State = syncstate.UpToDate
	}
	return status
}

func (s SyncStatus) WithError(failure SyncError) SyncStatus {
	s.State = syncstate.Error
	s.LastError = &failure
	return s
}

func (s SyncStatus) InConflict() SyncStatus {
	s.State = syncstate.Conflict
	return s
}

func (s SyncStatus) SyncedAt(at time.Time) SyncStatus {
	s.LastSyncAt = &at
	return s
}

func (s SyncStatus) WithSharedSettings(shared SyncSettings) SyncStatus {
	if shared.IsEmpty() || !shared.Differs(s.Settings.Shared()) {
		return s
	}
	s.Notice = fmt.Sprintf(sharedSettingsNotice, shared.Repository, shared.Branch, shared.Subfolder)
	return s
}

func (s SyncStatus) LogValue() any {
	return map[string]any{"state": s.State, "localChanges": s.LocalChanges, "ahead": s.Ahead, "behind": s.Behind, "lastError": s.LastError}
}
