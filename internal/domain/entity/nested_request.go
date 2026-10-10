package entity

type NestedRequest struct {
	Response  *Response
	RequestID string
	Error     string
	Request   EffectiveRequest
}

func (n NestedRequest) HasFailed() bool {
	return n.Error != "" || (n.Response != nil && n.Response.Error != nil)
}

func (n NestedRequest) IsRunRequest() bool {
	return IsSavedRequestID(n.RequestID)
}
