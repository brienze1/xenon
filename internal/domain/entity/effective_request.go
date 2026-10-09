package entity

import "strings"

type EffectiveRequest struct {
	Auth                Auth
	Body                Body
	Method              string
	URL                 string
	Headers             []KeyValue
	AutoHeaders         []AutoHeader
	DisabledAutoHeaders []string
	Settings            RequestSettings
}

func (e EffectiveRequest) HasHeader(name string) bool {
	for _, header := range e.Headers {
		if header.HasKey(name) {
			return true
		}
	}
	return false
}

func (e EffectiveRequest) IsAutoHeaderDisabled(name string) bool {
	for _, disabled := range e.DisabledAutoHeaders {
		if strings.EqualFold(strings.TrimSpace(disabled), name) {
			return true
		}
	}
	return false
}

func (e EffectiveRequest) WithAutoHeaders(headers []AutoHeader) EffectiveRequest {
	e.AutoHeaders = headers
	return e
}

func (e EffectiveRequest) UsesAutoHeader(name string) bool {
	for _, header := range e.AutoHeaders {
		if strings.EqualFold(header.Key, name) {
			return header.IsActive()
		}
	}
	return false
}

func (e EffectiveRequest) SendHeaders() []KeyValue {
	headers := append([]KeyValue{}, e.Headers...)
	for _, header := range e.AutoHeaders {
		if header.IsSent() {
			headers = append(headers, NewKeyValue(header.Key, header.Value))
		}
	}
	return headers
}

func (e EffectiveRequest) TimeoutText() string {
	return e.Settings.TimeoutText()
}

func (e EffectiveRequest) LogValue() any {
	return map[string]any{"method": e.Method, "url": e.URL}
}
