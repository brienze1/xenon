package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPEnvEdit struct {
	ConfirmBeforeSend *bool   `json:"confirm_before_send"`
	Name              *string `json:"name"`
	Color             *string `json:"color"`
	Index             *int    `json:"index"`
	Id                string  `json:"id"`
}

func (m *MCPEnvEdit) EnvironmentName() string {
	if m.Name == nil {
		return ""
	}
	return *m.Name
}

func (m *MCPEnvEdit) EditsSettings() bool {
	return m.Color != nil || m.ConfirmBeforeSend != nil
}

func (m *MCPEnvEdit) WithoutName() MCPEnvEdit {
	edit := *m
	edit.Name = nil
	return edit
}

func (m *MCPEnvEdit) CreateEnvironment(localId string) CreateEnvironment {
	return NewCreateEnvironment(localId, m.EnvironmentName())
}

func (m *MCPEnvEdit) EnvironmentReference(localId string) EnvironmentReference {
	return NewEnvironmentReference(localId, m.Id)
}

func (m *MCPEnvEdit) UpdateEnvironment(localId string, stored entity.Environment) UpdateEnvironment {
	return NewUpdateEnvironment(localId, stored, m)
}

func (m *MCPEnvEdit) ReorderEnvironment(localId string) ReorderEnvironment {
	if m.Index == nil {
		return NewReorderEnvironment(localId, m.Id, -1)
	}
	return NewReorderEnvironment(localId, m.Id, *m.Index)
}
