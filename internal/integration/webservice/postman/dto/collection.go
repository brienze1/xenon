package dto

import (
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
)

var collectionKeys = []string{"info", "item", "auth", "event", "variable"}

type Collection struct {
	Raw      map[string]any `json:"-"`
	Auth     *Auth          `json:"auth"`
	Info     Info           `json:"info"`
	Items    []Item         `json:"item"`
	Event    []Event        `json:"event"`
	Variable []KeyValue     `json:"variable"`
}

func (c *Collection) UnmarshalJSON(data []byte) error {
	type plain Collection
	if err := json.Unmarshal(data, (*plain)(c)); err != nil {
		return err
	}
	c.Raw = decodeRaw(data)
	return nil
}

func (c *Collection) ToEntity(conversion *Conversion) entity.ImportNode {
	collection := entity.NewCollection(conversion.newId(), c.Info.Name)
	collection.Description = c.Info.Description.Content
	collection.Auth = c.Auth.ToEntity()
	collection.Scripts = scriptsToEntity(c.Event)
	for index := range c.Variable {
		collection.Variables = append(collection.Variables, c.Variable[index].ToVariableRow())
	}
	collection.Postman = c.Passthrough()
	conversion.checkAuth(c.Auth, collection.ID, collection.Name)
	children := make([]entity.ImportNode, 0, len(c.Items))
	for index := range c.Items {
		children = append(children, c.Items[index].ToEntity(conversion, collection.Name))
	}
	return entity.NewContainerImportNode(collection, children)
}

func (c *Collection) Passthrough() map[string]any {
	fields := without(c.Raw, collectionKeys...)
	info := withEntry(without(c.Info.Raw, "name", "description"), "description", c.Info.Description.Passthrough())
	fields = withEntry(fields, "info", info)
	fields = withEntry(fields, "event", eventsPassthrough(c.Event))
	return withEntry(fields, "auth", c.Auth.Passthrough())
}
