package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ItemSettings struct {
	Auth           Auth       `json:"auth"`
	Scripts        *Scripts   `json:"scripts"`
	InheritScripts *bool      `json:"inherit_scripts,omitempty"`
	Id             string     `json:"id"`
	Kind           string     `json:"kind"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Headers        []KeyValue `json:"headers"`
}

func NewItemSettings(container entity.Collection) ItemSettings {
	settings := ItemSettings{
		Id:          container.ID,
		Kind:        container.Kind.String(),
		Name:        container.Name,
		Description: container.Description,
		Auth:        NewAuth(container.Auth),
		Headers:     NewKeyValues(container.Headers),
		Scripts:     NewScripts(container.Scripts),
	}
	if container.IsFolder() {
		inherit := container.InheritScripts
		settings.InheritScripts = &inherit
	}
	return settings
}

func (i ItemSettings) LogValue() any {
	return map[string]any{"id": i.Id, "kind": i.Kind}
}
