package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type Request struct {
	Settings       Settings       `yaml:"settings" json:"settings"`
	Postman        map[string]any `yaml:"postman" json:"-"`
	Bruno          map[string]any `yaml:"bruno" json:"-"`
	Kind           string         `yaml:"kind" json:"kind"`
	Id             string         `yaml:"id" json:"id"`
	Name           string         `yaml:"name" json:"name"`
	Description    string         `yaml:"description" json:"description"`
	Method         string         `yaml:"method" json:"method"`
	Url            string         `yaml:"url" json:"url"`
	Body           Body           `yaml:"body" json:"body"`
	Scripts        Scripts        `yaml:"scripts" json:"-"`
	InheritScripts *bool          `yaml:"inheritScripts" json:"-"`
	Params         []KeyValue     `yaml:"params" json:"params"`
	PathVariables  []PathVariable `yaml:"pathVariables" json:"path_variables"`
	Headers        []KeyValue     `yaml:"headers" json:"headers"`
	AutoHeaders    AutoHeaders    `yaml:"autoHeaders" json:"auto_headers"`
	Examples       []Example      `yaml:"examples" json:"-"`
	Auth           Auth           `yaml:"auth" json:"auth"`
	Xenon          int            `yaml:"xenon" json:"xenon"`
	ScriptsEdited  bool           `yaml:"-" json:"-"`
}

func NewRequest(request entity.Request) Request {
	model := Request{
		Xenon:          formatVersion,
		Kind:           itemkind.Request.String(),
		Id:             request.ID,
		Name:           request.Name,
		Description:    request.Description,
		Method:         request.MethodOrDefault(),
		Url:            request.URL,
		Params:         newKeyValues(request.Params),
		PathVariables:  make([]PathVariable, 0, len(request.PathVariables)),
		Headers:        newKeyValues(request.Headers),
		AutoHeaders:    NewAutoHeaders(request.DisabledAutoHeaders),
		Auth:           NewAuth(request.Auth),
		Body:           NewBody(request.Body),
		Settings:       NewSettings(request.Settings),
		Scripts:        NewScripts(request.Scripts),
		InheritScripts: &request.InheritScripts,
		ScriptsEdited:  request.ScriptsEdited,
		Examples:       newExamples(request.Examples),
		Postman:        request.Postman,
		Bruno:          request.Bruno,
	}
	for _, variable := range request.PathVariables {
		model.PathVariables = append(model.PathVariables, NewPathVariable(variable))
	}
	return model
}

func (r *Request) ToEntity() entity.Request {
	request := entity.NewRequest(r.Id, r.Name)
	request.Description = r.Description
	request.URL = r.Url
	if r.Method != "" {
		request.Method = r.Method
	}
	request.Method = request.NormalizedMethod()
	request.Params = keyValuesToEntity(r.Params)
	for index := range r.PathVariables {
		request.PathVariables = append(request.PathVariables, r.PathVariables[index].ToEntity())
	}
	request.Headers = keyValuesToEntity(r.Headers)
	request.DisabledAutoHeaders = r.AutoHeaders.ToEntity()
	request.Auth = r.Auth.ToEntity()
	request.Body = r.Body.ToEntity()
	request.Settings = r.Settings.ToEntity()
	request.Scripts = r.Scripts.ToEntity()
	request.InheritScripts = inheritsScripts(r.InheritScripts)
	request.Examples = examplesToEntity(r.Examples)
	request.Postman = sourceFieldsToEntity(r.Postman)
	request.Bruno = sourceFieldsToEntity(r.Bruno)
	return request
}

func (r *Request) ToStoredEntity(id, revision string) entity.Request {
	request := r.ToEntity()
	request.ID = id
	request.Revision = revision
	return request
}

func (r *Request) Fields() []yamlnode.Field {
	return r.fields(optionalExamples(r.Examples))
}

func (r *Request) FieldsWithExamples() []yamlnode.Field {
	return r.fields(examplesField(r.Examples))
}

func (r *Request) fields(examples []yamlnode.Field) []yamlnode.Field {
	fields := append(headerFields(itemkind.Request.String(), r.Id, r.Name, r.Description),
		yamlnode.String("method", r.Method),
		yamlnode.String("url", r.Url),
		yamlnode.List("params", keyValueNodes(r.Params)),
		yamlnode.List("pathVariables", pathVariableNodes(r.PathVariables)),
		yamlnode.List("headers", keyValueNodes(r.Headers)),
		yamlnode.Object("autoHeaders", yamlnode.StringList("disabled", r.AutoHeaders.ToEntity())),
		r.Auth.Field(),
		r.Body.Field(),
		r.Settings.Field(),
		scriptsField(r.Scripts, r.ScriptsEdited),
		scriptsReserved(yamlnode.Reserved("inheritScripts", inheritScriptsNode(r.InheritScripts)), r.ScriptsEdited),
	)
	fields = append(fields, examples...)
	fields = append(fields, sourceFieldsReserved(postmanKey, r.Postman))
	return append(fields, optionalSourceFields(brunoKey, r.Bruno)...)
}
