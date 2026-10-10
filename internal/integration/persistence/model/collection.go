package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type Collection struct {
	Auth        Auth          `yaml:"auth"`
	Kind        string        `yaml:"kind"`
	Id          string        `yaml:"id"`
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Order       []string      `yaml:"order"`
	Headers     []KeyValue    `yaml:"headers"`
	Variables   []VariableRow `yaml:"variables"`
	Xenon       int           `yaml:"xenon"`
}

func NewCollection(collection entity.Collection) Collection {
	return Collection{
		Xenon:       formatVersion,
		Kind:        collection.Kind.String(),
		Id:          collection.ID,
		Name:        collection.Name,
		Description: collection.Description,
		Order:       nonNil(collection.Order),
		Auth:        NewAuth(collection.Auth),
		Headers:     newKeyValues(collection.Headers),
		Variables:   newVariableRows(collection.Variables),
	}
}

func (c *Collection) ToEntity() entity.Collection {
	return entity.Collection{
		ID:          c.Id,
		Kind:        itemkind.GetItemKind(c.Kind),
		Name:        c.Name,
		Description: c.Description,
		Order:       nonNil(c.Order),
		Auth:        c.Auth.ToEntity(),
		Headers:     keyValuesToEntity(c.Headers),
		Variables:   variableRowsToEntity(c.Variables),
	}
}

func (c *Collection) ToContainerEntity(id string, kind enums.ItemKind) entity.Collection {
	container := c.ToEntity()
	container.ID = id
	container.Kind = kind
	return container
}

func (c *Collection) Fields() []yamlnode.Field {
	fields := append(headerFields(c.Kind, c.Id, c.Name, c.Description),
		yamlnode.StringList("order", nonNil(c.Order)),
		c.Auth.Field(),
		yamlnode.List("headers", keyValueNodes(c.Headers)),
	)
	if itemkind.Collection.Equals(c.Kind) {
		fields = append(fields, variablesField(c.Variables))
	}
	fields = append(fields, yamlnode.Reserved("scripts", emptyScriptsNode()))
	if itemkind.Folder.Equals(c.Kind) {
		fields = append(fields, yamlnode.Reserved("inheritScripts", inheritScriptsNode()))
	}
	return append(fields, yamlnode.Reserved("postman", yamlnode.EmptyMapping()))
}
