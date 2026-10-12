package entity

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
)

type SyncSession struct {
	StartedAt  time.Time
	ID         string
	LocalID    string
	LocalHead  string
	RemoteHead string
	Base       string
	Conflicts  []ConflictItem
	Replay     bool
}

func NewSyncSession(id, localId string, heads BranchHeads, replay bool, conflicts []ConflictItem, startedAt time.Time) SyncSession {
	seen := map[string]bool{}
	unique := make([]ConflictItem, 0, len(conflicts))
	for _, conflict := range conflicts {
		if !seen[conflict.ItemID] {
			seen[conflict.ItemID] = true
			unique = append(unique, conflict)
		}
	}
	return SyncSession{
		ID:         id,
		LocalID:    localId,
		LocalHead:  heads.Local,
		RemoteHead: heads.Remote,
		Base:       heads.Base,
		Replay:     replay,
		Conflicts:  unique,
		StartedAt:  startedAt,
	}
}

func (s SyncSession) Heads() BranchHeads {
	return BranchHeads{Local: s.LocalHead, Remote: s.RemoteHead, Base: s.Base}
}

func (s SyncSession) Conflict(itemId string) (ConflictItem, bool) {
	for _, conflict := range s.Conflicts {
		if conflict.ItemID == itemId {
			return conflict, true
		}
	}
	return ConflictItem{}, false
}

func (s SyncSession) WithConflict(resolved ConflictItem) SyncSession {
	conflicts := make([]ConflictItem, 0, len(s.Conflicts))
	for _, conflict := range s.Conflicts {
		if conflict.ItemID == resolved.ItemID {
			conflict = resolved
		}
		conflicts = append(conflicts, conflict)
	}
	s.Conflicts = conflicts
	return s
}

func (s SyncSession) IsResolved() bool {
	for _, conflict := range s.Conflicts {
		if !conflict.IsResolved() {
			return false
		}
	}
	return true
}

func (s SyncSession) StartedFrom(localHead string) bool {
	return s.LocalHead == localHead
}

func (s SyncSession) CommittedConflicts() []ConflictItem {
	return s.conflictsWhere(func(conflict ConflictItem) bool { return !conflict.Uncommitted })
}

func (s SyncSession) UncommittedConflicts() []ConflictItem {
	return s.conflictsWhere(func(conflict ConflictItem) bool { return conflict.Uncommitted })
}

func (s SyncSession) FieldResolutions() map[string]map[string]enums.ConflictResolution {
	resolutions := map[string]map[string]enums.ConflictResolution{}
	for _, conflict := range s.Conflicts {
		if conflict.IsFieldMerge() {
			resolutions[conflict.Path] = conflict.FieldResolutions()
		}
	}
	return resolutions
}

func (s SyncSession) conflictsWhere(matches func(ConflictItem) bool) []ConflictItem {
	selected := make([]ConflictItem, 0, len(s.Conflicts))
	for _, conflict := range s.Conflicts {
		if matches(conflict) {
			selected = append(selected, conflict)
		}
	}
	return selected
}

func (s SyncSession) LogValue() any {
	return map[string]any{"id": s.ID, "localId": s.LocalID, "localHead": s.LocalHead, "remoteHead": s.RemoteHead, "conflicts": len(s.Conflicts), "replay": s.Replay}
}

// ResolvedTree applies the chosen side of every committed item conflict to tree; field merges are already in it.
func (s SyncSession) ResolvedTree(tree map[string]string) map[string]string {
	resolved := make(map[string]string, len(tree))
	for path, hash := range tree {
		resolved[path] = hash
	}
	for _, conflict := range s.CommittedConflicts() {
		if conflict.IsFieldMerge() {
			continue
		}
		delete(resolved, conflict.Path)
		if result := conflict.Result(); result != "" {
			resolved[conflict.Path] = result
		}
	}
	return resolved
}
