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

func (c *Collection) ToEntity(conversion *Conversion) entity.ImportCollection {
	collection := entity.NewCollection(conversion.newId(), c.Info.Name)
	collection.Description = c.Info.Description.Content
	collection.Auth = c.Auth.ToEntity()
	collection.Scripts = scriptsToEntity(c.Event)
	secrets := map[string]string{}
	for index := range c.Variable {
		row := c.Variable[index].ToVariableRow()
		if row.HasStraySecretValue() {
			secrets[row.Key] = row.InitialValue
		}
		collection.Variables = append(collection.Variables, row.WithoutSecretValue())
	}
	collection.Postman = c.Passthrough()
	conversion.checkAuth(c.Auth, collection.ID, collection.Name)
	children := make([]entity.ImportNode, 0, len(c.Items))
	for index := range c.Items {
		children = append(children, c.Items[index].ToEntity(conversion, collection.Name))
	}
	return entity.NewImportCollection(entity.NewContainerImportNode(collection, children), conversion.Warnings(), secrets)
}

func (c *Collection) Passthrough() map[string]any {
	fields := without(c.Raw, collectionKeys...)
	info := withEntry(without(c.Info.Raw, "name", "description"), "description", c.Info.Description.Passthrough())
	fields = withEntry(fields, "info", info)
	fields = withEntry(fields, "event", eventsPassthrough(c.Event))
	return withEntry(fields, "auth", c.Auth.Passthrough())
}

func NewCollection(root entity.ExportNode) Collection {
	container := root.Container
	passthrough := map[string]any(container.Postman)
	info, rest := asMap(passthrough["info"]), without(passthrough, collectionKeys...)
	if container.IsFolder() {
		info = map[string]any{"description": passthrough["description"]}
		if id, found := passthrough["id"]; found {
			info["_postman_id"] = id
		}
		rest = without(passthrough, append(itemKeys, "id", "variable")...)
	}
	id, found := info["_postman_id"]
	if !found {
		id = container.ID
	}
	variables := make([]KeyValue, 0, len(container.Variables))
	for _, row := range container.Variables {
		variables = append(variables, NewVariable(row))
	}
	return Collection{
		Info:     NewInfo(container.Name, container.Description, id, info),
		Items:    NewItems(root.Children, !hasPostmanFields(root)),
		Auth:     NewAuth(container.Auth, passthrough["auth"]),
		Event:    NewEvents(container.Scripts, root.InheritedScripts, passthrough["event"]),
		Variable: variables,
		Raw:      rest,
	}
}

func (c Collection) MarshalJSON() ([]byte, error) {
	collection := newObject().set("info", c.Info).set("item", nonNil(c.Items))
	if c.Auth != nil {
		collection.set("auth", c.Auth)
	}
	collection.setIf(len(c.Event) > 0, "event", c.Event).setIf(len(c.Variable) > 0, "variable", c.Variable)
	return collection.merged(c.Raw).MarshalJSON()
}

func hasPostmanFields(root entity.ExportNode) bool {
	for _, node := range root.Nodes() {
		if !node.Container.Postman.IsEmpty() || (node.IsRequest() && !node.Request.Postman.IsEmpty()) {
			return true
		}
	}
	return false
}
