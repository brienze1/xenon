package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type envEditArguments struct {
	ConfirmBeforeSend *bool   `json:"confirm_before_send"`
	Name              *string `json:"name"`
	Color             *string `json:"color"`
	Id                string  `json:"id"`
}

type envActiveArguments struct {
	Id *string `json:"id"`
}

type envVariableArguments struct {
	Scope       string `json:"scope"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Secret      bool   `json:"secret"`
}

type envVariablesArguments struct {
	Scope string                 `json:"scope"`
	Vars  []variableRowArguments `json:"vars"`
}

type envActiveView struct {
	ActiveId string `json:"active_id"`
}

type envScopeVariables struct {
	Scope string         `json:"scope"`
	Vars  []variableLine `json:"vars"`
}

func (t *tools) envEdit(ctx context.Context, call Call) (Result, error) {
	result, err := t.editEnvironments(ctx, call)
	if err != nil {
		return Result{}, err
	}
	result.Changes = []dto.AIChange{{Kind: aichangekind.Variables.String()}}
	return result, nil
}

func (t *tools) editEnvironments(ctx context.Context, call Call) (Result, error) {
	switch call.Action {
	case "create":
		return t.createEnvironment(ctx, call)
	case "update":
		return t.updateEnvironment(ctx, call)
	case "duplicate":
		return t.duplicateEnvironment(ctx, call)
	case "set_active":
		return t.setActiveEnvironment(ctx, call)
	case "add_var":
		return t.addVariable(ctx, call)
	case "set_vars":
		return t.setVariables(ctx, call)
	case "set_current":
		return t.setCurrentValue(ctx, call)
	default:
		return t.resetOrPersistVariable(ctx, call)
	}
}

func (t *tools) createEnvironment(ctx context.Context, call Call) (Result, error) {
	var arguments envEditArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	name := ""
	if arguments.Name != nil {
		name = *arguments.Name
	}
	created, err := t.controllers.EnvironmentManage.Create(ctx, dto.CreateEnvironment{LocalId: call.LocalId, Name: name})
	if err != nil {
		return Result{}, err
	}
	if arguments.Color != nil || arguments.ConfirmBeforeSend != nil {
		arguments.Id, arguments.Name = created.Id, nil
		return t.savedEnvironment(ctx, call, *created, arguments)
	}
	return environmentResult(*created), nil
}

func (t *tools) updateEnvironment(ctx context.Context, call Call) (Result, error) {
	var arguments envEditArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	stored, err := t.controllers.EnvironmentManage.Get(ctx, dto.EnvironmentReference{LocalId: call.LocalId, Id: arguments.Id})
	if err != nil {
		return Result{}, err
	}
	return t.savedEnvironment(ctx, call, *stored, arguments)
}

func (t *tools) savedEnvironment(ctx context.Context, call Call, stored dto.Environment, arguments envEditArguments) (Result, error) {
	update := dto.UpdateEnvironment{LocalId: call.LocalId, Id: stored.Id, Name: stored.Name, Description: stored.Description, Color: stored.Color, ConfirmBeforeSend: stored.ConfirmBeforeSend}
	if arguments.Name != nil {
		update.Name = *arguments.Name
	}
	if arguments.Color != nil {
		update.Color = *arguments.Color
	}
	if arguments.ConfirmBeforeSend != nil {
		update.ConfirmBeforeSend = *arguments.ConfirmBeforeSend
	}
	updated, err := t.controllers.EnvironmentManage.Update(ctx, update)
	if err != nil {
		return Result{}, err
	}
	return environmentResult(*updated), nil
}

func (t *tools) duplicateEnvironment(ctx context.Context, call Call) (Result, error) {
	var arguments envEditArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	duplicated, err := t.controllers.EnvironmentManage.Duplicate(ctx, dto.EnvironmentReference{LocalId: call.LocalId, Id: arguments.Id})
	if err != nil {
		return Result{}, err
	}
	return environmentResult(*duplicated), nil
}

func (t *tools) setActiveEnvironment(ctx context.Context, call Call) (Result, error) {
	var arguments envActiveArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	id := ""
	if arguments.Id != nil {
		id = *arguments.Id
	}
	active, err := t.controllers.EnvironmentSetActive.Handle(ctx, dto.SetActiveEnvironment{LocalId: call.LocalId, Id: id})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: envActiveView{ActiveId: active.ActiveId}, Target: dto.AIActivityTarget{Id: active.ActiveId}}, nil
}

func (t *tools) addVariable(ctx context.Context, call Call) (Result, error) {
	var arguments envVariableArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	row, err := t.controllers.VariableTable.AddVariable(ctx, dto.AddVariable{LocalId: call.LocalId, Scope: arguments.Scope, Key: arguments.Key, Value: arguments.Value, Description: arguments.Description, Secret: arguments.Secret})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newVariableLine(*row), Target: variableTarget(arguments.Scope, arguments.Key)}, nil
}

func (t *tools) setVariables(ctx context.Context, call Call) (Result, error) {
	var arguments envVariablesArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	table, err := t.saveVariableRows(ctx, call, arguments.Scope, arguments.Vars)
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: envScopeVariables{Scope: table.Scope, Vars: newVariableLines(table.Rows)}, Target: dto.AIActivityTarget{Path: arguments.Scope}}, nil
}

func (t *tools) setCurrentValue(ctx context.Context, call Call) (Result, error) {
	var arguments envVariableArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	row, err := t.controllers.VariableTable.SetCurrent(ctx, dto.SetCurrentValue{LocalId: call.LocalId, Scope: arguments.Scope, Key: arguments.Key, Value: arguments.Value})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newVariableLine(*row), Target: variableTarget(arguments.Scope, arguments.Key)}, nil
}

func (t *tools) resetOrPersistVariable(ctx context.Context, call Call) (Result, error) {
	var arguments envVariableArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	keys := dto.VariableKeys{LocalId: call.LocalId, Scope: arguments.Scope, Keys: []string{arguments.Key}}
	change := t.controllers.VariableTable.Reset
	if call.Action == "persist" {
		change = t.controllers.VariableTable.Persist
	}
	table, err := change(ctx, keys)
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: envScopeVariables{Scope: table.Scope, Vars: newVariableLines(table.Rows)}, Target: variableTarget(arguments.Scope, arguments.Key)}, nil
}

func environmentResult(environment dto.Environment) Result {
	return Result{
		Payload: envLine{Id: environment.Id, Name: environment.Name, Color: environment.Color, ConfirmBeforeSend: environment.ConfirmBeforeSend},
		Target:  dto.AIActivityTarget{Id: environment.Id, Name: environment.Name},
	}
}

func variableTarget(scope, key string) dto.AIActivityTarget {
	return dto.AIActivityTarget{Name: key, Path: scope}
}
