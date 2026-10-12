package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SyncSession struct {
	ID        string             `json:"id"`
	LocalId   string             `json:"local_id"`
	Conflicts []SyncConflictItem `json:"conflicts"`
}

func NewSyncSession(session entity.SyncSession) SyncSession {
	conflicts := make([]SyncConflictItem, 0, len(session.Conflicts))
	for _, conflict := range session.Conflicts {
		conflicts = append(conflicts, NewSyncConflictItem(conflict))
	}
	return SyncSession{ID: session.ID, LocalId: session.LocalID, Conflicts: conflicts}
}

func (s SyncSession) LogValue() any {
	return map[string]any{"id": s.ID, "localId": s.LocalId, "conflicts": len(s.Conflicts)}
}
