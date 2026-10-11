package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SaveItemSettings struct {
	Auth           Auth       `json:"auth"`
	Scripts        *Scripts   `json:"scripts,omitempty"`
	InheritScripts *bool      `json:"inherit_scripts,omitempty"`
	LocalId        string     `json:"local_id" validate:"required,not_blank"`
	Id             string     `json:"id" validate:"required,not_blank"`
	Description    string     `json:"description"`
	Headers        []KeyValue `json:"headers"`
}

func (s *SaveItemSettings) ToEntity() entity.Collection {
	settings := entity.Collection{ID: s.Id, Description: s.Description, Auth: s.Auth.ToEntity(), Headers: keyValuesToEntity(s.Headers)}
	if s.Scripts != nil {
		settings.Scripts = s.Scripts.ToEntity()
		settings.InheritScripts = inheritsScripts(s.InheritScripts)
		settings.ScriptsEdited = true
	}
	return settings
}

func (s SaveItemSettings) LogValue() any {
	return map[string]any{"local_id": s.LocalId, "id": s.Id}
}
