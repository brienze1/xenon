package entity

type NestedRequest struct {
	Response  *Response
	RequestID string
	Error     string
	Scopes    ScriptScopes
	Request   EffectiveRequest
	Skipped   bool
}

func NewSentNestedRequest(request EffectiveRequest, response Response) NestedRequest {
	return NestedRequest{Request: request, Response: &response}
}

func NewRejectedNestedRequest(message string) NestedRequest {
	return NestedRequest{Error: message}
}

// NewRunNestedRequest is the outcome of pm.execution.runRequest: the saved request's send with the scopes its scripts left.
func NewRunNestedRequest(requestId string, execution SendExecution) NestedRequest {
	nested := NestedRequest{RequestID: requestId, Response: execution.Response, Scopes: execution.Scopes, Skipped: !execution.ShouldSend()}
	if execution.SentRequest != nil {
		nested.Request = *execution.SentRequest
	}
	return nested
}

func NewMissingRunNestedRequest(requestId, message string) NestedRequest {
	return NestedRequest{RequestID: requestId, Error: message}
}

func (n NestedRequest) HasFailed() bool {
	return n.Error != "" || (n.Response != nil && n.Response.Error != nil)
}

func (n NestedRequest) FailureMessage() string {
	if n.Error != "" {
		return n.Error
	}
	if n.Response != nil && n.Response.Error != nil {
		return n.Response.Error.Message
	}
	return ""
}

func (n NestedRequest) IsRunRequest() bool {
	return n.RequestID != ""
}

func (n NestedRequest) LogValue() any {
	value := map[string]any{"method": n.Request.Method, "url": n.Request.URL, "error": n.Error, "requestId": n.RequestID, "skipped": n.Skipped}
	if n.Response != nil {
		value["status_code"] = n.Response.StatusCode
	}
	return value
}
