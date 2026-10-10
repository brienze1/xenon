package entity

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums/httpmethod"
)

var methodToken = regexp.MustCompile(`^[A-Z][A-Z0-9_-]{0,31}$`)

type Request struct {
	Auth                Auth
	Body                Body
	ID                  string
	Name                string
	Description         string
	Method              string
	URL                 string
	Params              []KeyValue
	PathVariables       []PathVariable
	Headers             []KeyValue
	Revision            string
	DisabledAutoHeaders []string
	Settings            RequestSettings
}

func NewRequest(id, name string) Request {
	return Request{
		ID:                  id,
		Name:                strings.TrimSpace(name),
		Method:              httpmethod.Get.String(),
		Params:              []KeyValue{},
		PathVariables:       []PathVariable{},
		Headers:             []KeyValue{},
		DisabledAutoHeaders: []string{},
		Auth:                DefaultAuth(),
		Body:                DefaultBody(),
		Settings:            DefaultRequestSettings(),
	}
}

func (r Request) HasValidName() bool {
	return isValidItemName(r.Name)
}

func (r Request) NormalizedMethod() string {
	return strings.ToUpper(strings.TrimSpace(r.Method))
}

func (r Request) MethodOrDefault() string {
	if method := r.NormalizedMethod(); method != "" {
		return method
	}
	return httpmethod.Get.String()
}

func (r Request) HasID() bool {
	return strings.TrimSpace(r.ID) != ""
}

func (r Request) HasValidMethod() bool {
	return methodToken.MatchString(r.NormalizedMethod())
}

func (r Request) CopyAs(id, name string) Request {
	r.ID = id
	r.Name = strings.TrimSpace(name)
	r.Revision = ""
	return r
}

func (r Request) WithFileRefsRelativeTo(workspaceRoot string) Request {
	r.Body = r.Body.RelativeTo(workspaceRoot)
	return r
}

func (r Request) MapText(transform func(string) string) Request {
	r.URL = transform(r.URL)
	r.Params = mapKeyValues(r.Params, transform)
	pathVariables := make([]PathVariable, 0, len(r.PathVariables))
	for _, variable := range r.PathVariables {
		pathVariables = append(pathVariables, variable.MapText(transform))
	}
	r.PathVariables = pathVariables
	r.Headers = mapKeyValues(r.Headers, transform)
	r.Auth = r.Auth.MapText(transform)
	r.Body = r.Body.MapText(transform)
	return r
}

// TextFields lists the texts variables are resolved in, named after the request DTO: the URL, params, path variables, headers, the active auth type and the active body mode.
func (r Request) TextFields() []TextField {
	fields := []TextField{NewTextField("url", r.URL)}
	fields = append(fields, keyValueFields("params", r.Params)...)
	for index, variable := range r.PathVariables {
		fields = append(fields, NewTextField(fieldName("path_variables", strconv.Itoa(index), "value"), variable.Value))
	}
	fields = append(fields, keyValueFields("headers", r.Headers)...)
	fields = append(fields, r.Auth.TextFields("auth")...)
	return append(fields, r.Body.TextFields("body")...)
}

func (r Request) LogValue() any {
	return map[string]any{"id": r.ID, "method": r.Method, "url": r.URL}
}

func (r Request) IsAutoHeaderDisabled(name string) bool {
	for _, disabled := range r.DisabledAutoHeaders {
		if strings.EqualFold(disabled, name) {
			return true
		}
	}
	return false
}

func (r Request) PathVariableKeys() []string {
	keys := make([]string, 0, len(r.PathVariables))
	for _, variable := range r.PathVariables {
		keys = append(keys, variable.Key)
	}
	return keys
}

func (r Request) ShouldEncodeUrl() bool {
	return r.Settings.EncodeUrl
}
