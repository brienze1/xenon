package dto

type MCPEnvInspect struct {
	Text      string `json:"text" description:"Text whose {{variables}} to inspect when no request_id is given."`
	RequestId string `json:"request_id" description:"Saved request whose variables to inspect." validated_as:"id"`
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
