package dto

type MCPBinaryBody struct {
	ContentType string `json:"content_type,omitempty"`
	Size        int64  `json:"size"`
	Binary      bool   `json:"binary"`
}

func NewMCPBinaryBody(result SendResult) MCPBinaryBody {
	return MCPBinaryBody{Binary: true, ContentType: result.ContentType, Size: result.SizeBytes}
}
