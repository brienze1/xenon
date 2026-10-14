package dto

type MCPTreeExampleResponse struct {
	Body         *string    `json:"body,omitempty"`
	StatusText   string     `json:"status_text,omitempty"`
	ContentType  string     `json:"content_type,omitempty"`
	BodyEncoding string     `json:"body_encoding,omitempty"`
	Headers      []KeyValue `json:"headers,omitempty"`
	Status       int        `json:"status,omitempty"`
	Size         int        `json:"size,omitempty"`
	Binary       bool       `json:"binary,omitempty"`
}

func NewMCPTreeExampleResponse(response ExampleResponse) MCPTreeExampleResponse {
	return MCPTreeExampleResponse{
		Body:         response.Body,
		Status:       response.Code,
		StatusText:   response.Status,
		ContentType:  response.ContentType,
		BodyEncoding: response.BodyEncoding,
		Headers:      response.Headers,
		Size:         response.SizeBytes,
		Binary:       response.IsBinary,
	}
}
