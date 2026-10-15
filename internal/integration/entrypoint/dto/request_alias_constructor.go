package dto

import "github.com/brienze1/xenon/internal/domain/entity"

func NewRequestAlias(alias entity.RequestAlias) RequestAlias {
	return RequestAlias{Id: alias.ID, Name: alias.Name, Request: NewRequest(alias.Version)}
}

func NewRequestAliases(aliases []entity.RequestAlias) []RequestAlias {
	result := make([]RequestAlias, 0, len(aliases))
	for _, alias := range aliases {
		result = append(result, NewRequestAlias(alias))
	}
	return result
}
