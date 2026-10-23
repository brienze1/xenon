package dto

import "github.com/brienze1/xenon/internal/domain/entity"

func NewRequestAlias(alias entity.RequestAlias) RequestAlias {
	return RequestAlias{Id: alias.ID, Name: alias.Name, EnvironmentIds: append([]string{}, alias.EnvironmentIDs...), Request: NewRequest(alias.Version)}
}

func NewRequestAliases(aliases []entity.RequestAlias) []RequestAlias {
	result := make([]RequestAlias, 0, len(aliases))
	for _, alias := range aliases {
		result = append(result, NewRequestAlias(alias))
	}
	return result
}

func NewEmptyRequestAlias(id string) RequestAlias {
	return RequestAlias{Id: id}
}

func NewEditedRequestAlias(alias RequestAlias, edit *MCPTreeEdit) RequestAlias {
	return RequestAlias{Id: edit.Id, Name: overridden(edit.Name, alias.Name), EnvironmentIds: alias.EnvironmentIds, Request: NewEditedRequest(alias.Request, edit)}
}
