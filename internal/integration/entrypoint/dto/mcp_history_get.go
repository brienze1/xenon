package dto

type MCPHistoryGet struct {
	Id string `json:"id"`
}

func (m *MCPHistoryGet) HistoryReference(localId string) HistoryReference {
	return NewHistoryReference(localId, m.Id)
}
