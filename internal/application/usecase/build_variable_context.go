package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

// BuildVariableContext orders the scopes environment, collection (the root ancestor), globals; nil environment is "No environment".
func BuildVariableContext(environment *entity.Environment, ancestors []entity.Collection, globals []entity.VariableRow, local entity.LocalVariables, vault entity.Vault) entity.VariableContext {
	layers := make([]entity.VariableLayer, 0, 3)
	if environment != nil {
		layers = append(layers, variableLayer(variablescope.Environment, environment.ScopeRef(), environment.Variables, local, vault))
	}
	if len(ancestors) > 0 {
		collection := ancestors[0]
		layers = append(layers, variableLayer(variablescope.Collection, collection.ScopeRef(), collection.OwnVariables(), local, vault))
	}
	layers = append(layers, variableLayer(variablescope.Globals, entity.NewGlobalsScopeRef(), globals, local, vault))
	return entity.NewVariableContext(layers...)
}

func variableLayer(kind enums.VariableScope, scope entity.ScopeRef, rows []entity.VariableRow, local entity.LocalVariables, vault entity.Vault) entity.VariableLayer {
	return entity.NewVariableLayer(kind, rows, local.CurrentValues(scope), vault.SecretsFor(scope), vault.IsLocked()).WithRef(scope)
}
