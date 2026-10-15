package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

const aliasesKey = "aliases"

type RequestAlias struct {
	Settings       Settings       `yaml:"settings"`
	Id             string         `yaml:"id"`
	Name           string         `yaml:"name"`
	Method         string         `yaml:"method"`
	Url            string         `yaml:"url"`
	Body           Body           `yaml:"body"`
	Scripts        Scripts        `yaml:"scripts"`
	InheritScripts *bool          `yaml:"inheritScripts"`
	Params         []KeyValue     `yaml:"params"`
	PathVariables  []PathVariable `yaml:"pathVariables"`
	Headers        []KeyValue     `yaml:"headers"`
	AutoHeaders    AutoHeaders    `yaml:"autoHeaders"`
	Auth           Auth           `yaml:"auth"`
}

func NewRequestAlias(alias entity.RequestAlias) RequestAlias {
	version := alias.Version
	model := RequestAlias{
		Id:             alias.ID,
		Name:           alias.Name,
		Method:         version.MethodOrDefault(),
		Url:            version.URL,
		Params:         newKeyValues(version.Params),
		PathVariables:  make([]PathVariable, 0, len(version.PathVariables)),
		Headers:        newKeyValues(version.Headers),
		AutoHeaders:    NewAutoHeaders(version.DisabledAutoHeaders),
		Auth:           NewAuth(version.Auth),
		Body:           NewBody(version.Body),
		Settings:       NewSettings(version.Settings),
		Scripts:        NewScripts(version.Scripts),
		InheritScripts: &version.InheritScripts,
	}
	for _, variable := range version.PathVariables {
		model.PathVariables = append(model.PathVariables, NewPathVariable(variable))
	}
	return model
}

func (a *RequestAlias) ToEntity() entity.RequestAlias {
	version := entity.NewRequest("", "")
	version.URL = a.Url
	if a.Method != "" {
		version.Method = a.Method
	}
	version.Method = version.NormalizedMethod()
	version.Params = keyValuesToEntity(a.Params)
	for index := range a.PathVariables {
		version.PathVariables = append(version.PathVariables, a.PathVariables[index].ToEntity())
	}
	version.Headers = keyValuesToEntity(a.Headers)
	version.DisabledAutoHeaders = a.AutoHeaders.ToEntity()
	version.Auth = a.Auth.ToEntity()
	version.Body = a.Body.ToEntity()
	version.Settings = a.Settings.ToEntity()
	version.Scripts = a.Scripts.ToEntity()
	version.InheritScripts = inheritsScripts(a.InheritScripts)
	return entity.RequestAlias{ID: a.Id, Name: a.Name, Version: version}
}

func (a *RequestAlias) Node() *yaml.Node {
	return yamlnode.MappingNode(
		yamlnode.String("id", a.Id),
		yamlnode.String("name", a.Name),
		yamlnode.String("method", a.Method),
		yamlnode.String("url", a.Url),
		yamlnode.List("params", keyValueNodes(a.Params)),
		yamlnode.List("pathVariables", pathVariableNodes(a.PathVariables)),
		yamlnode.List("headers", keyValueNodes(a.Headers)),
		yamlnode.Object("autoHeaders", yamlnode.StringList("disabled", a.AutoHeaders.ToEntity())),
		a.Auth.Field(),
		a.Body.Field(),
		a.Settings.Field(),
		yamlnode.Field{Key: scriptsKey, Value: a.Scripts.Node()},
		yamlnode.Bool("inheritScripts", inheritsScripts(a.InheritScripts)),
	)
}

func newRequestAliases(aliases []entity.RequestAlias) []RequestAlias {
	models := make([]RequestAlias, 0, len(aliases))
	for _, alias := range aliases {
		models = append(models, NewRequestAlias(alias))
	}
	return models
}

func requestAliasesToEntity(models []RequestAlias) []entity.RequestAlias {
	if len(models) == 0 {
		return nil
	}
	aliases := make([]entity.RequestAlias, 0, len(models))
	for index := range models {
		aliases = append(aliases, models[index].ToEntity())
	}
	return aliases
}

func optionalAliases(models []RequestAlias) []yamlnode.Field {
	if len(models) == 0 {
		return []yamlnode.Field{}
	}
	return []yamlnode.Field{yamlnode.Reserved(aliasesKey, aliasNodes(models))}
}

func aliasesField(models []RequestAlias) []yamlnode.Field {
	if len(models) == 0 {
		return []yamlnode.Field{yamlnode.Absent(aliasesKey)}
	}
	return []yamlnode.Field{{Key: aliasesKey, Value: aliasNodes(models)}}
}

func aliasNodes(models []RequestAlias) *yaml.Node {
	nodes := make([]*yaml.Node, 0, len(models))
	for index := range models {
		nodes = append(nodes, models[index].Node())
	}
	return yamlnode.List(aliasesKey, nodes).Value
}
