package mcp

import (
	"context"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

const (
	scopeSeparator   = ":"
	globalsScope     = "globals"
	environmentScope = "env"
	collectionScope  = "collection"
	globalsName      = "Globals"
)

type envScopeArguments struct {
	Scope string `json:"scope"`
}

type envInspectArguments struct {
	Text      string `json:"text"`
	RequestId string `json:"request_id"`
}

type envLine struct {
	Id                string `json:"id"`
	Name              string `json:"name"`
	Color             string `json:"color"`
	Active            bool   `json:"active,omitempty"`
	ConfirmBeforeSend bool   `json:"confirm_before_send,omitempty"`
}

type envScopeView struct {
	Name string         `json:"name"`
	Vars []variableLine `json:"vars"`
}

type envReferenceLine struct {
	Name  string `json:"name"`
	Scope string `json:"scope"`
	Value string `json:"value"`
}

type envInspection struct {
	Refs []envReferenceLine `json:"refs"`
}

func newEnvLine(environment dto.EnvironmentSummary, activeId string) envLine {
	return envLine{Id: environment.Id, Name: environment.Name, Color: environment.Color, Active: environment.Id == activeId, ConfirmBeforeSend: environment.ConfirmBeforeSend}
}

func newEnvReferenceLine(usage dto.VariableUsage) envReferenceLine {
	line := envReferenceLine{Name: usage.Name, Scope: usage.Scope, Value: usage.DisplayValue}
	if usage.Secret {
		line.Value = entity.MaskedValue
	}
	return line
}

func (t *tools) env(ctx context.Context, call Call) (Result, error) {
	switch call.Action {
	case "get":
		return t.envGet(ctx, call)
	case "inspect":
		return t.envInspect(ctx, call)
	default:
		return t.envList(ctx, call)
	}
}

func (t *tools) envList(ctx context.Context, call Call) (Result, error) {
	environments, err := t.controllers.EnvironmentList.Handle(ctx, dto.ListEnvironments{LocalId: call.LocalId})
	if err != nil {
		return Result{}, err
	}
	lines := make([]envLine, 0, len(environments.Environments))
	for _, environment := range environments.Environments {
		lines = append(lines, newEnvLine(environment, environments.ActiveId))
	}
	return Result{Payload: List[envLine]{Items: lines}}, nil
}

func (t *tools) envGet(ctx context.Context, call Call) (Result, error) {
	var arguments envScopeArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	table, err := t.controllers.VariableTable.GetTable(ctx, dto.VariableTableReference{LocalId: call.LocalId, Scope: arguments.Scope})
	if err != nil {
		return Result{}, err
	}
	name := t.scopeName(ctx, call.LocalId, table.Scope)
	return Result{Payload: envScopeView{Name: name, Vars: newVariableLines(table.Rows)}, Target: dto.AIActivityTarget{Name: name, Path: table.Scope}}, nil
}

func (t *tools) scopeName(ctx context.Context, localId, scope string) string {
	kind, id, _ := strings.Cut(scope, scopeSeparator)
	switch kind {
	case environmentScope:
		if environment, err := t.controllers.EnvironmentManage.Get(ctx, dto.EnvironmentReference{LocalId: localId, Id: id}); err == nil {
			return environment.Name
		}
	case collectionScope:
		if item, err := t.findWorkspaceItem(ctx, localId, id); err == nil {
			return item.Name
		}
	case globalsScope:
		return globalsName
	}
	return scope
}

func (t *tools) envInspect(ctx context.Context, call Call) (Result, error) {
	var arguments envInspectArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	request, target, err := t.inspectedRequest(ctx, call, arguments)
	if err != nil {
		return Result{}, err
	}
	inspection, err := t.controllers.VariableInspect.Inspect(ctx, dto.InspectVariables{LocalId: call.LocalId, Request: request})
	if err != nil {
		return Result{}, err
	}
	refs := make([]envReferenceLine, 0, len(inspection.Variables))
	for _, usage := range inspection.Variables {
		refs = append(refs, newEnvReferenceLine(usage))
	}
	return Result{Payload: envInspection{Refs: refs}, Target: target}, nil
}

func (t *tools) inspectedRequest(ctx context.Context, call Call, arguments envInspectArguments) (dto.Request, dto.AIActivityTarget, error) {
	if arguments.RequestId == "" {
		return dto.Request{Url: arguments.Text}, dto.AIActivityTarget{}, nil
	}
	item, err := t.findWorkspaceItem(ctx, call.LocalId, arguments.RequestId)
	if err != nil {
		return dto.Request{}, dto.AIActivityTarget{}, err
	}
	request, err := t.controllers.RequestGet.Get(ctx, dto.OpenRequest{LocalId: call.LocalId, Id: item.Id})
	if err != nil {
		return dto.Request{}, dto.AIActivityTarget{}, err
	}
	return *request, item.target(), nil
}
