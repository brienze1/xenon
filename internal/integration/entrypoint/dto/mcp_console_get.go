package dto

type MCPConsoleGet struct {
	Id string `json:"id"`
}

func (m *MCPConsoleGet) ConsoleEntryReference() ConsoleEntryReference {
	return NewConsoleEntryReference(m.Id)
}
