package dto

type MCPTreeItem struct {
	Id string `json:"id"`
}

func (m *MCPTreeItem) HasId() bool {
	return m.Id != ""
}

func (m *MCPTreeItem) ItemId() string {
	return m.Id
}
