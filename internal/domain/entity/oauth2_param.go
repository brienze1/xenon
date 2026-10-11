package entity

import "strings"

const (
	OAuth2SendAsUrl    = "request_url"
	OAuth2SendAsBody   = "request_body"
	OAuth2SendAsHeader = "request_header"
)

type OAuth2Param struct {
	Key     string
	Value   string
	SendAs  string
	Enabled bool
}

func (o OAuth2Param) IsActive() bool {
	return o.Enabled && strings.TrimSpace(o.Key) != ""
}

func (o OAuth2Param) SendsInUrl() bool {
	return o.SendAs == OAuth2SendAsUrl
}

func (o OAuth2Param) SendsInHeader() bool {
	return o.SendAs == OAuth2SendAsHeader
}

func (o OAuth2Param) SendsInBody() bool {
	return !o.SendsInUrl() && !o.SendsInHeader()
}

func (o OAuth2Param) MapText(transform func(string) string) OAuth2Param {
	o.Key = transform(o.Key)
	o.Value = transform(o.Value)
	return o
}
