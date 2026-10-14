package dto

type MCPEnvActivation struct {
	Id *string `json:"id"`
}

func (m *MCPEnvActivation) SetActiveEnvironment(localId string) SetActiveEnvironment {
	id := ""
	if m.Id != nil {
		id = *m.Id
	}
	return NewSetActiveEnvironment(localId, id)
}
