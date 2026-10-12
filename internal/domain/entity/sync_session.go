package entity

import "time"

type SyncSession struct {
	StartedAt  time.Time
	ID         string
	LocalHead  string
	RemoteHead string
	Base       string
	Conflicts  []ConflictItem
}

func NewSyncSession(id string, heads BranchHeads, conflicts []ConflictItem, startedAt time.Time) SyncSession {
	seen := map[string]bool{}
	unique := make([]ConflictItem, 0, len(conflicts))
	for _, conflict := range conflicts {
		if !seen[conflict.ItemID] {
			seen[conflict.ItemID] = true
			unique = append(unique, conflict)
		}
	}
	return SyncSession{ID: id, LocalHead: heads.Local, RemoteHead: heads.Remote, Base: heads.Base, Conflicts: unique, StartedAt: startedAt}
}
