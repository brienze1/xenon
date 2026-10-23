package dto

type MCPRunRow struct {
	RequestId string `json:"request_id" description:"Request id."`
	AliasId   string `json:"alias_id" description:"Alias id; empty for the Default."`
}

func (m *MCPRunRow) RunRow() RunRow {
	return RunRow{RequestId: m.RequestId, AliasId: m.AliasId, Selected: true}
}
