package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type Collection struct {
	Postman        map[string]any `yaml:"postman"`
	Bruno          map[string]any `yaml:"bruno"`
	Kind           string         `yaml:"kind"`
	Id             string         `yaml:"id"`
	Name           string         `yaml:"name"`
	Description    string         `yaml:"description"`
	Scripts        Scripts        `yaml:"scripts"`
	InheritScripts *bool          `yaml:"inheritScripts"`
	Order          []string       `yaml:"order"`
	Headers        []KeyValue     `yaml:"headers"`
	Variables      []VariableRow  `yaml:"variables"`
	Auth           Auth           `yaml:"auth"`
	Xenon          int            `yaml:"xenon"`
	ScriptsEdited  bool           `yaml:"-"`
}

func NewCollection(collection entity.Collection) Collection {
	return Collection{
		Xenon:          formatVersion,
		Kind:           collection.Kind.String(),
		Id:             collection.ID,
		Name:           collection.Name,
		Description:    collection.Description,
		Order:          nonNil(collection.Order),
		Auth:           NewAuth(collection.Auth),
		Headers:        newKeyValues(collection.Headers),
		Variables:      newVariableRows(collection.Variables),
		Scripts:        NewScripts(collection.Scripts),
		InheritScripts: &collection.InheritScripts,
		ScriptsEdited:  collection.ScriptsEdited,
		Postman:        collection.Postman,
		Bruno:          collection.Bruno,
	}
}

func (c *Collection) ToEntity() entity.Collection {
	return entity.Collection{
		ID:             c.Id,
		Kind:           itemkind.GetItemKind(c.Kind),
		Name:           c.Name,
		Description:    c.Description,
		Order:          nonNil(c.Order),
		Auth:           c.Auth.ToEntity(),
		Headers:        keyValuesToEntity(c.Headers),
		Variables:      variableRowsToEntity(c.Variables),
		Scripts:        c.Scripts.ToEntity(),
		InheritScripts: inheritsScripts(c.InheritScripts),
		Postman:        sourceFieldsToEntity(c.Postman),
		Bruno:          sourceFieldsToEntity(c.Bruno),
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
	fields = append(fields, scriptsField(c.Scripts, c.ScriptsEdited))
	if itemkind.Folder.Equals(c.Kind) {
		fields = append(fields, scriptsReserved(yamlnode.Reserved("inheritScripts", inheritScriptsNode(c.InheritScripts)), c.ScriptsEdited))
	}
	fields = append(fields, sourceFieldsReserved(postmanKey, c.Postman))
	return append(fields, optionalSourceFields(brunoKey, c.Bruno)...)
}

// ImportedFields writes scripts and auth details even when the file already has them.
func (c *Collection) ImportedFields() []yamlnode.Field {
	fields := c.Fields()
	for index, field := range fields {
		if field.Key == scriptsKey || field.Key == authKey {
			fields[index] = overwritten(field)
		}
	}
	return fields
}

func overwritten(field yamlnode.Field) yamlnode.Field {
	field.KeepExisting = false
	nested := make([]yamlnode.Field, 0, len(field.Fields))
	for _, child := range field.Fields {
		nested = append(nested, overwritten(child))
	}
	if field.Fields != nil {
		field.Fields = nested
	}
	return field
}
