package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type Collection struct {
	Config    map[string]any
	Name      string
	Docs      string
	File      string
	Headers   []Pair
	Variables []Pair
	Auth      Auth
	Scripts   Scripts
}

func NewCollection(name string) Collection {
	return Collection{Name: name}
}

func (c *Collection) ToEntity(conversion *Conversion) entity.Collection {
	collection := entity.NewCollection(conversion.newId(), c.Name)
	collection.Description = c.Docs
	collection.Auth = c.Auth.ToEntity()
	collection.Headers = keyValues(c.Headers)
	for _, pair := range c.Variables {
		row := entity.NewVariableRow(pair.Key, pair.Value)
		row.Enabled = pair.Enabled
		collection.Variables = append(collection.Variables, row)
	}
	if len(c.Config) > 0 {
		collection.Bruno = c.Config
	}
	conversion.checkAuth(c.Auth, collection.ID, collection.Name)
	conversion.addScripts(c.Scripts, collection.ID, collection.Name, c.File)
	return collection
}
