package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type Request struct {
	Auth           Auth            `json:"auth"`
	Settings       RequestSettings `json:"settings"`
	Id             string          `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Method         string          `json:"method"`
	Url            string          `json:"url"`
	Revision       string          `json:"revision"`
	Body           Body            `json:"body"`
	Scripts        *Scripts        `json:"scripts,omitempty"`
	InheritScripts *bool           `json:"inherit_scripts,omitempty"`
	AutoHeaders    AutoHeaders     `json:"auto_headers"`
	Params         []KeyValue      `json:"params"`
	PathVariables  []PathVariable  `json:"path_variables"`
	Headers        []KeyValue      `json:"headers"`
}

func NewRequest(request entity.Request) Request {
	result := Request{
		Id:             request.ID,
		Name:           request.Name,
		Description:    request.Description,
		Method:         request.NormalizedMethod(),
		Url:            request.URL,
		Revision:       request.Revision,
		Params:         NewKeyValues(request.Params),
		PathVariables:  make([]PathVariable, 0, len(request.PathVariables)),
		Headers:        NewKeyValues(request.Headers),
		AutoHeaders:    NewAutoHeaders(request.DisabledAutoHeaders),
		Auth:           NewAuth(request.Auth),
		Body:           NewBody(request.Body),
		Settings:       NewRequestSettings(request.Settings),
		Scripts:        NewScripts(request.Scripts),
		InheritScripts: &request.InheritScripts,
	}
	for _, variable := range request.PathVariables {
		result.PathVariables = append(result.PathVariables, NewPathVariable(variable))
	}
	return result
}

func (r *Request) ToEntity() entity.Request {
	request := entity.NewRequest(r.Id, r.Name)
	request.Description = r.Description
	if r.Method != "" {
		request.Method = r.Method
	}
	request.Method = request.NormalizedMethod()
	request.URL = r.Url
	request.Params = keyValuesToEntity(r.Params)
	for index := range r.PathVariables {
		request.PathVariables = append(request.PathVariables, r.PathVariables[index].ToEntity())
	}
	request.Headers = keyValuesToEntity(r.Headers)
	request.DisabledAutoHeaders = append(request.DisabledAutoHeaders, r.AutoHeaders.Disabled...)
	request.Auth = r.Auth.ToEntity()
	request.Body = r.Body.ToEntity()
	request.Settings = r.Settings.ToEntity()
	if r.Scripts != nil {
		request.Scripts = r.Scripts.ToEntity()
		request.InheritScripts = inheritsScripts(r.InheritScripts)
		request.ScriptsEdited = true
	}
	return request
}

func (r Request) LogValue() any {
	return map[string]any{"id": r.Id, "method": r.Method, "url": r.Url}
}
