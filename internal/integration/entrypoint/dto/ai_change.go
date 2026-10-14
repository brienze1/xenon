package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
)

type AIChange struct {
	Kind      string `json:"kind" validate:"required,oneof=tree request variables history cookies workspaces"`
	RequestId string `json:"request_id,omitempty"`
	Revision  string `json:"revision,omitempty"`
}

func NewAIChange(kind enums.AIChangeKind) AIChange {
	return AIChange{Kind: kind.String()}
}

func NewRequestAIChange(requestId, revision string) AIChange {
	return AIChange{Kind: aichangekind.Request.String(), RequestId: requestId, Revision: revision}
}

func (a *AIChange) ToEntity() entity.AIChange {
	return entity.AIChange{Kind: aichangekind.GetAIChangeKind(a.Kind), RequestID: a.RequestId, Revision: a.Revision}
}
