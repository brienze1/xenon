package dto

type MCPTreeTokenOwner struct {
	OwnerId string `json:"owner_id" description:"Id of the item that owns the OAuth token." required_for:"token"`
}

func (m *MCPTreeTokenOwner) HasOwner() bool {
	return m.OwnerId != ""
}

func (m *MCPTreeTokenOwner) OwnerItemId() string {
	return m.OwnerId
}

func (m *MCPTreeTokenOwner) OAuthTokenReference(localId string) OAuthTokenReference {
	return NewOAuthTokenReference(localId, m.OwnerId)
}
