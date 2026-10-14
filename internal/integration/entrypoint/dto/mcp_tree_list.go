package dto

import "encoding/json"

const (
	mcpTreeAllDepth       = "all"
	mcpTreeUnlimitedDepth = -1
	mcpTreeDefaultDepth   = 1
	mcpTreeNullDepth      = "null"
)

type MCPTreeList struct {
	ParentId string          `json:"parent_id"`
	Cursor   string          `json:"cursor"`
	Depth    json.RawMessage `json:"depth"`
}

// ListDepth reads depth as absent/null (1 level), "all" (every level) or an integer of 1 or more; false means invalid.
func (m *MCPTreeList) ListDepth() (int, bool) {
	if len(m.Depth) == 0 || string(m.Depth) == mcpTreeNullDepth {
		return mcpTreeDefaultDepth, true
	}
	var all string
	if json.Unmarshal(m.Depth, &all) == nil && all == mcpTreeAllDepth {
		return mcpTreeUnlimitedDepth, true
	}
	var depth int
	if err := json.Unmarshal(m.Depth, &depth); err != nil || depth < mcpTreeDefaultDepth {
		return 0, false
	}
	return depth, true
}

func (m *MCPTreeList) HasParent() bool {
	return m.ParentId != ""
}

func (m *MCPTreeList) ParentItemId() string {
	return m.ParentId
}

func (m *MCPTreeList) PageCursor() string {
	return m.Cursor
}
