package entity

type NestedRequest struct {
	Response  *Response
	RequestID string
	Error     string
	Request   EffectiveRequest
}

func NewSentNestedRequest(request EffectiveRequest, response Response) NestedRequest {
	return NestedRequest{Request: request, Response: &response}
}

func NewRejectedNestedRequest(message string) NestedRequest {
	return NestedRequest{Error: message}
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
	return IsSavedRequestID(n.RequestID)
}

func (n NestedRequest) LogValue() any {
	value := map[string]any{"method": n.Request.Method, "url": n.Request.URL, "error": n.Error}
	if n.Response != nil {
		value["status_code"] = n.Response.StatusCode
	}
	return value
}
