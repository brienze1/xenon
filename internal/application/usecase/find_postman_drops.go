package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	dropInheritScripts  = "Postman always runs the parent scripts"
	dropHeaders         = "Postman has no collection or folder headers; they are copied onto every request below"
	dropBruno           = "Bruno-only fields have no Postman equivalent"
	dropBrunoAuth       = "This Bruno auth type has no Postman equivalent; it is exported as No Auth"
	dropOAuth2Fields    = "The token to use and auto-refresh are Xenon-only OAuth 2.0 settings"
	dropTimeout         = "Postman has no per-request timeout"
	dropBinaryExample   = "Postman v2.1 cannot hold binary example bodies; the body is exported empty"
	dropVaultReferences = "Xenon resolves {{vault:…}} from the workspace secrets; Postman looks them up in its own vault"
	dropColor           = "Postman environments have no colour"
	dropConfirm         = "Postman has no confirm-before-send"
	dropSecretVariable  = "Postman collection v2.1 has no secret variables; it is exported as an empty string variable"
)

// FindPostmanDrops lists the Xenon-only data a Postman export of root loses; the root's own inherit-scripts toggle is honoured by the export and not listed.
func FindPostmanDrops(root entity.ExportNode) []entity.ExportDrop {
	drops := make([]entity.ExportDrop, 0)
	for index, node := range root.Nodes() {
		if node.IsRequest() {
			drops = append(drops, requestDrops(node, index == 0)...)
			continue
		}
		drops = append(drops, containerDrops(node, index == 0)...)
	}
	return drops
}

func FindPostmanEnvironmentDrops(environment entity.Environment, path string) []entity.ExportDrop {
	drops := make([]entity.ExportDrop, 0)
	if environment.HasColor() {
		drops = append(drops, entity.NewExportDrop(path, "color", dropColor))
	}
	if environment.ConfirmBeforeSend {
		drops = append(drops, entity.NewExportDrop(path, "confirmBeforeSend", dropConfirm))
	}
	if !environment.Bruno.IsEmpty() {
		drops = append(drops, entity.NewExportDrop(path, "bruno", dropBruno))
	}
	return drops
}

func FindPostmanSecretVariableDrops(rows []entity.VariableRow, path string) []entity.ExportDrop {
	drops := make([]entity.ExportDrop, 0)
	for _, row := range rows {
		if row.Secret {
			drops = append(drops, entity.NewExportDrop(path, "variables."+row.Key, dropSecretVariable))
		}
	}
	return drops
}

func containerDrops(node entity.ExportNode, isRoot bool) []entity.ExportDrop {
	container := node.Container
	drops := make([]entity.ExportDrop, 0)
	if !isRoot && container.ScriptOwner().CutsInheritance() {
		drops = append(drops, entity.NewExportDrop(node.Path, "inheritScripts", dropInheritScripts))
	}
	if node.HasContainerHeaders() {
		drops = append(drops, entity.NewExportDrop(node.Path, "headers", dropHeaders))
	}
	return append(drops, sharedDrops(node.Path, container.Auth, container.Bruno, container.Auth.TextFields("auth"))...)
}

func requestDrops(node entity.ExportNode, isRoot bool) []entity.ExportDrop {
	request := *node.Request
	drops := make([]entity.ExportDrop, 0)
	if !isRoot && !request.InheritScripts {
		drops = append(drops, entity.NewExportDrop(node.Path, "inheritScripts", dropInheritScripts))
	}
	if request.Settings.HasTimeout() {
		drops = append(drops, entity.NewExportDrop(node.Path, "settings.timeoutMs", dropTimeout))
	}
	for _, example := range request.Examples {
		if example.Response.IsBinary() {
			drops = append(drops, entity.NewExportDrop(node.Path, "examples."+example.Name+".body", dropBinaryExample))
		}
	}
	return append(drops, sharedDrops(node.Path, request.Auth, request.Bruno, request.TextFields())...)
}

func sharedDrops(path string, auth entity.Auth, bruno entity.SourceFields, fields []entity.TextField) []entity.ExportDrop {
	drops := make([]entity.ExportDrop, 0)
	if auth.IsBrunoOnly() {
		drops = append(drops, entity.NewExportDrop(path, "auth", dropBrunoAuth))
	}
	if auth.HasXenonOnlyOAuth2Fields() {
		drops = append(drops, entity.NewExportDrop(path, "auth.oauth2.useToken, auth.oauth2.autoRefresh", dropOAuth2Fields))
	}
	if !bruno.IsEmpty() {
		drops = append(drops, entity.NewExportDrop(path, "bruno", dropBruno))
	}
	for _, field := range fields {
		if field.HasVaultReference() {
			drops = append(drops, entity.NewExportDrop(path, field.Name, dropVaultReferences))
		}
	}
	return drops
}
