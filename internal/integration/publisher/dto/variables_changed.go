package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const VariablesChangedEvent = "variables:changed"

type VariablesChanged struct {
	LocalId string   `json:"local_id"`
	Reason  string   `json:"reason,omitempty"`
	Scopes  []string `json:"scopes"`
}

func NewVariablesChanged(localId string, scopes []entity.ScopeRef, reason string) VariablesChanged {
	names := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		names = append(names, scope.String())
	}
	return VariablesChanged{LocalId: localId, Scopes: names, Reason: reason}
}
