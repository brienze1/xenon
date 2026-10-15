package dto

const aliasKind = "alias"

func (m *MCPTreeEdit) EditsAlias() bool {
	return m.Kind == aliasKind
}

func (m *MCPTreeEdit) CreateAlias(localId, requestId string) CreateAlias {
	return NewCreateAlias(localId, requestId, m.ItemName())
}

func (m *MCPTreeEdit) SaveAlias(localId string, stored Request) SaveAlias {
	return NewSaveAlias(localId, stored.Id, stored.Revision, NewEditedRequestAlias(stored.AliasOrEmpty(m.Id), m), true)
}
