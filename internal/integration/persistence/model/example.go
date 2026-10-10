package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

const examplesKey = "examples"

type Example struct {
	Postman  map[string]any  `yaml:"postman"`
	Id       string          `yaml:"id"`
	Name     string          `yaml:"name"`
	Request  ExampleRequest  `yaml:"request"`
	Response ExampleResponse `yaml:"response"`
}

func NewExample(example entity.SavedExample) Example {
	return Example{
		Id:       example.ID,
		Name:     example.Name,
		Request:  NewExampleRequest(example.Request),
		Response: NewExampleResponse(example.Response),
		Postman:  example.Postman,
	}
}

func (e *Example) ToEntity() entity.SavedExample {
	return entity.SavedExample{
		ID:       e.Id,
		Name:     e.Name,
		Request:  e.Request.ToEntity(),
		Response: e.Response.ToEntity(),
		Postman:  sourceFieldsToEntity(e.Postman),
	}
}

func (e *Example) Node() *yaml.Node {
	return yamlnode.MappingNode(append([]yamlnode.Field{
		yamlnode.String("id", e.Id),
		yamlnode.String("name", e.Name),
		yamlnode.Object("request", e.Request.Fields()...),
		yamlnode.Object("response", e.Response.Fields()...),
	}, optionalSourceFields(postmanKey, e.Postman)...)...)
}

func newExamples(examples []entity.SavedExample) []Example {
	models := make([]Example, 0, len(examples))
	for _, example := range examples {
		models = append(models, NewExample(example))
	}
	return models
}

func examplesToEntity(models []Example) []entity.SavedExample {
	if len(models) == 0 {
		return nil
	}
	examples := make([]entity.SavedExample, 0, len(models))
	for index := range models {
		examples = append(examples, models[index].ToEntity())
	}
	return examples
}

func optionalExamples(models []Example) []yamlnode.Field {
	if len(models) == 0 {
		return []yamlnode.Field{}
	}
	return []yamlnode.Field{yamlnode.Reserved(examplesKey, exampleNodes(models))}
}

func examplesField(models []Example) []yamlnode.Field {
	if len(models) == 0 {
		return []yamlnode.Field{yamlnode.Absent(examplesKey)}
	}
	return []yamlnode.Field{{Key: examplesKey, Value: exampleNodes(models)}}
}

func exampleNodes(models []Example) *yaml.Node {
	nodes := make([]*yaml.Node, 0, len(models))
	for index := range models {
		nodes = append(nodes, models[index].Node())
	}
	return yamlnode.List(examplesKey, nodes).Value
}
