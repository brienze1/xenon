package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AIChanges struct {
	LocalId string     `json:"local_id" validate:"required,not_blank"`
	Changes []AIChange `json:"changes" validate:"dive"`
}

func NewAIChanges(localId string, changes []AIChange) AIChanges {
	return AIChanges{LocalId: localId, Changes: changes}
}

func (a *AIChanges) ToEntity() []entity.AIChange {
	changes := make([]entity.AIChange, 0, len(a.Changes))
	for _, change := range a.Changes {
		changes = append(changes, change.ToEntity())
	}
	return changes
}

func (a *AIChanges) WorkspaceId() string {
	return a.LocalId
}
