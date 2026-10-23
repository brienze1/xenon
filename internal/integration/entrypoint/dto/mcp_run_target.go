package dto

type MCPRunTarget struct {
	Target string `json:"target" description:"Collection or folder path, like Shop / Orders."`
	ItemId string `json:"item_id" description:"Collection or folder id, instead of target." validated_as:"item_id"`
}

func (m *MCPRunTarget) IsAmbiguous() bool {
	return m.Target != "" && m.ItemId != ""
}

func (m *MCPRunTarget) IsMissing() bool {
	return m.Target == "" && m.ItemId == ""
}

func (m *MCPRunTarget) IsPath() bool {
	return m.ItemId == ""
}

func (m *MCPRunTarget) Path() string {
	return m.Target
}

func (m *MCPRunTarget) RunTarget() MCPRunTarget {
	return *m
}

func (m *MCPRunTarget) TargetItemId() string {
	return m.ItemId
}
