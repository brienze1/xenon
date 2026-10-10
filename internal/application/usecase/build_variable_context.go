package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

// BuildVariableContext orders the scopes environment, collection (the root ancestor), globals; nil environment is "No environment".
func BuildVariableContext(environment *entity.Environment, ancestors []entity.Collection, globals []entity.VariableRow, local entity.LocalVariables) entity.VariableContext {
	layers := make([]entity.VariableLayer, 0, 3)
	if environment != nil {
		layers = append(layers, entity.NewVariableLayer(variablescope.Environment, environment.Variables, local.CurrentValues(environment.ScopeRef()), nil))
	}
	if len(ancestors) > 0 {
		collection := ancestors[0]
		layers = append(layers, entity.NewVariableLayer(variablescope.Collection, collection.OwnVariables(), local.CurrentValues(collection.ScopeRef()), nil))
	}
	layers = append(layers, entity.NewVariableLayer(variablescope.Globals, globals, local.CurrentValues(entity.NewGlobalsScopeRef()), nil))
	return entity.NewVariableContext(layers...)
}
