package dto

type MCPConsoleGet struct {
	Id string `json:"id" description:"Console entry id, from console list." required_for:"get"`
}

func (m *MCPConsoleGet) ConsoleEntryReference() ConsoleEntryReference {
	return NewConsoleEntryReference(m.Id)
}
