package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
)

type AIChanges struct {
	LocalId string     `json:"local_id" validate:"required,not_blank"`
	Changes []AIChange `json:"changes" validate:"dive"`
}

type AIChange struct {
	Kind      string `json:"kind" validate:"required,oneof=tree request variables history cookies workspaces"`
	RequestId string `json:"request_id,omitempty"`
	Revision  string `json:"revision,omitempty"`
}

func (a *AIChanges) ToEntity() []entity.AIChange {
	changes := make([]entity.AIChange, 0, len(a.Changes))
	for _, change := range a.Changes {
		changes = append(changes, entity.AIChange{Kind: aichangekind.GetAIChangeKind(change.Kind), RequestID: change.RequestId, Revision: change.Revision})
	}
	return changes
}
