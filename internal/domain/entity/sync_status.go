package entity

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
)

type SyncStatus struct {
	LastSyncAt   *time.Time
	LastError    *SyncError
	State        enums.SyncState
	LocalChanges int
	Ahead        int
	Behind       int
}
