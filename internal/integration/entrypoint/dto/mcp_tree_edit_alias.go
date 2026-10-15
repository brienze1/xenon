package dto

const aliasKind = "alias"

// EditsAlias is a tree_edit call on an alias, which is addressed through its request rather than the tree.
func (m *MCPTreeEdit) EditsAlias() bool {
	return m.Kind == aliasKind
}

func (m *MCPTreeEdit) CreateAlias(localId, requestId string) CreateAlias {
	return CreateAlias{LocalId: localId, RequestId: requestId, Name: m.ItemName()}
}

// SaveAlias applies the edits to the stored alias; an alias the request does not have is sent as is, so saving it
// reports it missing.
func (m *MCPTreeEdit) SaveAlias(localId string, stored Request) SaveAlias {
	alias := stored.AliasOrEmpty(m.Id)
	name := alias.Name
	if m.Name != nil {
		name = *m.Name
	}
	edited := RequestAlias{Id: m.Id, Name: name, Request: NewEditedRequest(alias.Request, m)}
	return SaveAlias{LocalId: localId, RequestId: stored.Id, BaseRevision: stored.Revision, Alias: edited, Force: true}
}
