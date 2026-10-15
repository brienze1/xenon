package dto

type MCPTreeItem struct {
	Id string `json:"id" description:"Item id, from tree list or search." required_for:"get,settings" validated_as:"example_id"`
}

func (m *MCPTreeItem) HasId() bool {
	return m.Id != ""
}

func (m *MCPTreeItem) ItemId() string {
	return m.Id
}
