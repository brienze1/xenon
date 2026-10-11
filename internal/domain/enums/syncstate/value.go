package syncstate

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	UpToDate     enums.SyncState = "up_to_date"
	LocalChanges enums.SyncState = "local_changes"
	Ahead        enums.SyncState = "ahead"
	Behind       enums.SyncState = "behind"
	Diverged     enums.SyncState = "diverged"
	Conflict     enums.SyncState = "conflict"
	Error        enums.SyncState = "error"
	Syncing      enums.SyncState = "syncing"
	NotLinked    enums.SyncState = "not_linked"
	Unknown      enums.SyncState = "unknown"
)

func GetSyncState(s string) enums.SyncState {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "up_to_date":
		return UpToDate
	case "local_changes":
		return LocalChanges
	case "ahead":
		return Ahead
	case "behind":
		return Behind
	case "diverged":
		return Diverged
	case "conflict":
		return Conflict
	case "error":
		return Error
	case "syncing":
		return Syncing
	case "not_linked":
		return NotLinked
	default:
		return Unknown
	}
}
