package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPSendTarget struct {
	Request   *Request `json:"request" description:"Ad-hoc request to use; pass exactly one of request or request_id."`
	RequestId string   `json:"request_id" description:"Saved request id; pass exactly one of request or request_id." validated_as:"id"`
}

func (m *MCPSendTarget) IsAmbiguous() bool {
	return m.RequestId != "" && m.Request != nil
}

func (m *MCPSendTarget) IsAdHoc() bool {
	return m.Request != nil
}

func (m *MCPSendTarget) IsMissing() bool {
	return m.RequestId == ""
}

func (m *MCPSendTarget) SendTarget() MCPSendTarget {
	return *m
}

func (m *MCPSendTarget) AdHocRequest() Request {
	return *m.Request
}

func (m *MCPSendTarget) AdHocTarget() entity.AIActivityTarget {
	return entity.NewAIRequestTarget(m.Request.Method, m.Request.Url)
}

func (m *MCPSendTarget) RequestItemId() string {
	return m.RequestId
}
