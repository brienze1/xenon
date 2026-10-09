package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/httpmethod"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/persistence/yamlnode"
)

type Request struct {
	Auth          Auth           `yaml:"auth"`
	Body          Body           `yaml:"body"`
	Settings      Settings       `yaml:"settings"`
	Kind          string         `yaml:"kind"`
	Id            string         `yaml:"id"`
	Name          string         `yaml:"name"`
	Description   string         `yaml:"description"`
	Method        string         `yaml:"method"`
	Url           string         `yaml:"url"`
	Params        []KeyValue     `yaml:"params"`
	PathVariables []PathVariable `yaml:"pathVariables"`
	Headers       []KeyValue     `yaml:"headers"`
	AutoHeaders   AutoHeaders    `yaml:"autoHeaders"`
	Xenon         int            `yaml:"xenon"`
}

func NewRequest(request entity.Request) Request {
	model := Request{
		Xenon:         formatVersion,
		Kind:          itemkind.Request.String(),
		Id:            request.ID,
		Name:          request.Name,
		Description:   request.Description,
		Method:        request.NormalizedMethod(),
		Url:           request.URL,
		Params:        newKeyValues(request.Params),
		PathVariables: make([]PathVariable, 0, len(request.PathVariables)),
		Headers:       newKeyValues(request.Headers),
		AutoHeaders:   AutoHeaders{Disabled: nonNil(request.DisabledAutoHeaders)},
		Auth:          NewAuth(request.Auth),
		Body:          NewBody(request.Body),
		Settings:      NewSettings(request.Settings),
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
	request.DisabledAutoHeaders = nonNil(r.AutoHeaders.Disabled)
	request.Auth = r.Auth.ToEntity()
	request.Body = r.Body.ToEntity()
	request.Settings = r.Settings.ToEntity()
	return request
}

func (r *Request) Fields() []yamlnode.Field {
	return append(headerFields(itemkind.Request.String(), r.Id, r.Name, r.Description),
		yamlnode.String("method", r.methodOrDefault()),
		yamlnode.String("url", r.Url),
		yamlnode.List("params", keyValueNodes(r.Params)),
		yamlnode.List("pathVariables", r.pathVariableNodes()),
		yamlnode.List("headers", keyValueNodes(r.Headers)),
		yamlnode.Object("autoHeaders", yamlnode.StringList("disabled", nonNil(r.AutoHeaders.Disabled))),
		r.Auth.Field(),
		r.Body.Field(),
		r.Settings.Field(),
		yamlnode.Reserved("scripts", emptyScriptsNode()),
		yamlnode.Reserved("inheritScripts", inheritScriptsNode()),
		yamlnode.Reserved("postman", yamlnode.EmptyMapping()),
	)
}

func (r *Request) methodOrDefault() string {
	if r.Method == "" {
		return httpmethod.Get.String()
	}
	return r.Method
}

func (r *Request) pathVariableNodes() []*yaml.Node {
	nodes := make([]*yaml.Node, 0, len(r.PathVariables))
	for index := range r.PathVariables {
		nodes = append(nodes, r.PathVariables[index].Node())
	}
	return nodes
}
