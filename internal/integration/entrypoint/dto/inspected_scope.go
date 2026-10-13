package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type InspectedScope struct {
	Scope string `json:"scope"`
	Ref   string `json:"ref"`
}

func NewInspectedScope(scope entity.InspectedScope) InspectedScope {
	return InspectedScope{Scope: scope.Scope.String(), Ref: scope.Ref.String()}
}
