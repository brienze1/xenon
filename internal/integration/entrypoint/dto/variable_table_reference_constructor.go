package dto

import "github.com/brienze1/xenon/internal/domain/entity"

func NewVariableTableReference(localId string, scope entity.ScopeRef) VariableTableReference {
	return VariableTableReference{LocalId: localId, Scope: scope.String()}
}

func NewScopeVariableTableReference(localId, scope string) VariableTableReference {
	return VariableTableReference{LocalId: localId, Scope: scope}
}
