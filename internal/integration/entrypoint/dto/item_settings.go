package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ItemSettings struct {
	Auth        Auth       `json:"auth"`
	Id          string     `json:"id"`
	Kind        string     `json:"kind"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Headers     []KeyValue `json:"headers"`
}

func NewItemSettings(container entity.Collection) ItemSettings {
	return ItemSettings{
		Id:          container.ID,
		Kind:        container.Kind.String(),
		Name:        container.Name,
		Description: container.Description,
		Auth:        NewAuth(container.Auth),
		Headers:     NewKeyValues(container.Headers),
	}
}

func (i ItemSettings) LogValue() any {
	return map[string]any{"id": i.Id, "kind": i.Kind}
}
