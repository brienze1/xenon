package dto

type MCPEnvActivation struct {
	Id *string `json:"id" description:"Environment to make active; omit or leave empty to deactivate."`
}

func (m *MCPEnvActivation) SetActiveEnvironment(localId string) SetActiveEnvironment {
	id := ""
	if m.Id != nil {
		id = *m.Id
	}
	return NewSetActiveEnvironment(localId, id)
}
