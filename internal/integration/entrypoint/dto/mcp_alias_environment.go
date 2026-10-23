package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPAliasEnvironment struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func NewMCPAliasEnvironment(id, name string) MCPAliasEnvironment {
	return MCPAliasEnvironment{Id: id, Name: name}
}

func NewMCPAliasEnvironments(ids []string, environments []entity.Environment) []MCPAliasEnvironment {
	labels := entity.AliasEnvironmentLabels(ids, environments)
	result := make([]MCPAliasEnvironment, 0, len(ids))
	for index, id := range ids {
		result = append(result, NewMCPAliasEnvironment(id, labels[index]))
	}
	return result
}
