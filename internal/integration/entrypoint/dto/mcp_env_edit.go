package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPEnvEdit struct {
	ConfirmBeforeSend *bool   `json:"confirm_before_send" description:"Ask before sending with this environment." used_by:"create,update"`
	Name              *string `json:"name" description:"Environment name." used_by:"create,update" required_for:"create"`
	Color             *string `json:"color" description:"Environment color." used_by:"create,update" enum:"none,green,cyan,magenta,yellow,red"`
	Index             *int    `json:"index" description:"New 0-based position in the environment list." used_by:"reorder" required_for:"reorder"`
	Id                string  `json:"id" description:"Environment id." used_by:"update,duplicate,reorder" required_for:"update,duplicate,reorder"`
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
