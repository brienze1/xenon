package dto

type MCPHistoryGet struct {
	Id string `json:"id" description:"History entry id, from history list." required_for:"get"`
}

func (m *MCPHistoryGet) HistoryReference(localId string) HistoryReference {
	return NewHistoryReference(localId, m.Id)
}
