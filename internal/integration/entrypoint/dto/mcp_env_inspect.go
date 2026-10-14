package dto

type MCPEnvInspect struct {
	Text      string `json:"text"`
	RequestId string `json:"request_id"`
}

func (m *MCPEnvInspect) HasRequest() bool {
	return m.RequestId != ""
}

// Request inspects the raw text as the URL of an unsaved request.
func (m *MCPEnvInspect) Request() Request {
	return NewUnsavedRequest(m.Text)
}

func (m *MCPEnvInspect) RequestItemId() string {
	return m.RequestId
}
