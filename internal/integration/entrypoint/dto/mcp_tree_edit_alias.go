package dto

const aliasKind = "alias"

func (m *MCPTreeEdit) EditsAlias() bool {
	return m.Kind == aliasKind
}

func (m *MCPTreeEdit) CreateAlias(localId, requestId string) CreateAlias {
	return NewCreateAlias(localId, requestId, m.ItemName(), m.AliasEnvironmentIds())
}

// SetsAliasEnvironments is an edit naming the alias's environments; without them the link is kept.
func (m *MCPTreeEdit) SetsAliasEnvironments() bool {
	return m.EnvironmentIds != nil
}

func (m *MCPTreeEdit) AliasEnvironmentIds() []string {
	if m.EnvironmentIds == nil {
		return nil
	}
	return *m.EnvironmentIds
}

func (m *MCPTreeEdit) SetAliasEnvironments(localId string, saved Request) SetAliasEnvironments {
	return NewSetAliasEnvironments(localId, saved, m.Id, m.AliasEnvironmentIds())
}

func (m *MCPTreeEdit) SaveAlias(localId string, stored Request) SaveAlias {
	return NewSaveAlias(localId, stored.Id, stored.Revision, NewEditedRequestAlias(stored.AliasOrEmpty(m.Id), m), true)
}
