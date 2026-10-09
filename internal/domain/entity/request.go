package entity

import (
	"regexp"
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
