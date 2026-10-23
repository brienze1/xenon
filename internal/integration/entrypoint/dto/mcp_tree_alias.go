package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPTreeAlias struct {
	Scripts      *Scripts              `json:"scripts,omitempty"`
	Auth         Auth                  `json:"auth"`
	Body         Body                  `json:"body"`
	Id           string                `json:"id"`
	Name         string                `json:"name"`
	Method       string                `json:"method"`
	Url          string                `json:"url"`
	Environments []MCPAliasEnvironment `json:"environments"`
	Params       []KeyValue            `json:"params"`
	Headers      []KeyValue            `json:"headers"`
}

type MCPAliasEnvironment struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func NewMCPTreeAlias(alias RequestAlias, environments []entity.Environment) MCPTreeAlias {
	version := alias.Request
	return MCPTreeAlias{
		Id:           alias.Id,
		Name:         alias.Name,
		Method:       version.Method,
		Url:          version.Url,
		Environments: newMCPAliasEnvironments(alias.EnvironmentIds, environments),
		Params:       version.Params,
		Headers:      version.Headers,
		Body:         version.Body,
		Auth:         version.Auth,
		Scripts:      version.Scripts,
	}
}

func NewMCPTreeAliases(aliases []RequestAlias, environments []entity.Environment) []MCPTreeAlias {
	result := make([]MCPTreeAlias, 0, len(aliases))
	for _, alias := range aliases {
		result = append(result, NewMCPTreeAlias(alias, environments))
	}
	return result
}

func newMCPAliasEnvironments(ids []string, environments []entity.Environment) []MCPAliasEnvironment {
	labels := entity.AliasEnvironmentLabels(ids, environments)
	result := make([]MCPAliasEnvironment, 0, len(ids))
	for index, id := range ids {
		result = append(result, MCPAliasEnvironment{Id: id, Name: labels[index]})
	}
	return result
}
