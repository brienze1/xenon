package dto

import (
	"encoding/json"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/httpmethod"
)

var requestKeys = []string{"method", "url", "header", "body", "auth"}

type Request struct {
	Raw      map[string]any `json:"-"`
	Body     *Body          `json:"body"`
	Auth     *Auth          `json:"auth"`
	Method   string         `json:"method"`
	Url      Url            `json:"url"`
	Header   HeaderList     `json:"header"`
	IsString bool           `json:"-"`
}

func (r *Request) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err == nil {
		r.IsString = true
		r.Url = NewTextUrl(raw)
		return nil
	}
	type plain Request
	if err := json.Unmarshal(data, (*plain)(r)); err != nil {
		return err
	}
	r.Raw = decodeRaw(data)
	return nil
}

func (r *Request) method() string {
	if method := strings.ToUpper(strings.TrimSpace(r.Method)); method != "" {
		return method
	}
	return httpmethod.Get.String()
}

func (r *Request) ToEntity(request entity.Request) entity.Request {
	request.Method = r.method()
	request.URL = r.Url.String()
	request.Params = r.Url.Params()
	request.PathVariables = r.Url.PathVariables()
	request.Headers = r.Header.ToEntity()
	request.Body = r.Body.ToEntity()
	request.Auth = r.Auth.ToEntity()
	return request
}

func (r *Request) ToExampleRequest() entity.ExampleRequest {
	example := entity.NewExampleRequest()
	request := r.ToEntity(entity.NewRequest("", ""))
	example.Method = request.Method
	example.URL = request.URL
	example.Params = request.Params
	example.PathVariables = request.PathVariables
	example.Headers = request.Headers
	example.Body = request.Body
	example.Auth = request.Auth
	return example
}

func (r *Request) Passthrough() map[string]any {
	fields := without(r.Raw, requestKeys...)
	fields = withEntry(fields, "url", r.Url.Passthrough())
	fields = r.Header.Passthrough(fields)
	if r.Body != nil {
		fields = withEntry(fields, "body", r.Body.Passthrough())
	}
	return withEntry(fields, "auth", r.Auth.Passthrough())
}

func (r *Request) HasDisabledBody() bool {
	return r.Body != nil && r.Body.Disabled
}
