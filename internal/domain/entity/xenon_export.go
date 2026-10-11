package entity

import (
	"maps"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/exportshape"
)

type XenonExport struct {
	Node                   *ExportNode
	Variables              map[string][]VariableRow
	Kind                   enums.ExportItemKind
	Shape                  enums.ExportShape
	ItemID                 string
	Secrets                string
	CurrentValuesAsInitial bool
}

func NewXenonExport(request ExportRequest) XenonExport {
	return XenonExport{
		Kind:                   request.ItemKind,
		Shape:                  request.Shape,
		ItemID:                 request.ItemID,
		CurrentValuesAsInitial: request.CurrentValuesAsInitial,
		Variables:              map[string][]VariableRow{},
	}
}

func (x XenonExport) WithNode(node ExportNode) XenonExport {
	x.Node = &node
	return x
}

func (x XenonExport) WithVariables(scope ScopeRef, rows []VariableRow) XenonExport {
	variables := map[string][]VariableRow{}
	maps.Copy(variables, x.Variables)
	variables[scope.String()] = append([]VariableRow{}, rows...)
	x.Variables = variables
	return x
}

func (x XenonExport) WithCurrentValues(scopes []ExportScope, local LocalVariables) XenonExport {
	for _, scope := range scopes {
		if rows, changed := scope.CurrentValuesAsInitial(local); changed {
			x = x.WithVariables(scope.Ref, rows)
		}
	}
	return x
}

func (x XenonExport) WithSecrets(armored string) XenonExport {
	x.Secrets = armored
	return x
}

func (x XenonExport) VariablesFor(scope ScopeRef) ([]VariableRow, bool) {
	rows, found := x.Variables[scope.String()]
	return rows, found
}

func (x XenonExport) HasVariables() bool {
	return len(x.Variables) > 0
}

func (x XenonExport) HasSecrets() bool {
	return x.Secrets != ""
}

func (x XenonExport) IsArchive() bool {
	return x.Shape == exportshape.Zip
}

func (x XenonExport) LogValue() any {
	scopes := make([]string, 0, len(x.Variables))
	for scope := range x.Variables {
		scopes = append(scopes, scope)
	}
	return map[string]any{"kind": x.Kind, "shape": x.Shape, "itemId": x.ItemID, "currentValueScopes": scopes, "hasSecrets": x.HasSecrets()}
}
