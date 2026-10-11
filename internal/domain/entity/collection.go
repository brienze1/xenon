package entity

import (
	"slices"
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptownerkind"
)

type Collection struct {
	Postman        SourceFields
	Bruno          SourceFields
	Auth           Auth
	Scripts        Scripts
	ID             string
	Kind           enums.ItemKind
	Name           string
	Description    string
	Order          []string
	Headers        []KeyValue
	Variables      []VariableRow
	InheritScripts bool
	ScriptsEdited  bool
}

func NewCollection(id, name string) Collection {
	return newContainer(id, name, itemkind.Collection)
}

func NewFolder(id, name string) Collection {
	return newContainer(id, name, itemkind.Folder)
}

func newContainer(id, name string, kind enums.ItemKind) Collection {
	return Collection{
		ID:             id,
		Kind:           kind,
		Name:           strings.TrimSpace(name),
		Order:          []string{},
		Auth:           DefaultAuth(),
		Headers:        []KeyValue{},
		Variables:      []VariableRow{},
		InheritScripts: true,
	}
}

func (c Collection) IsFolder() bool {
	return c.Kind == itemkind.Folder
}

func (c Collection) HasValidName() bool {
	return isValidItemName(c.Name)
}

func (c Collection) WithSettings(settings Collection) Collection {
	c.Description = settings.Description
	c.Auth = settings.Auth
	c.Headers = append([]KeyValue{}, settings.Headers...)
	if settings.ScriptsEdited {
		c.Scripts = settings.Scripts
		c.InheritScripts = settings.InheritScripts
		c.ScriptsEdited = true
	}
	return c
}

func (c Collection) WithOrder(order []string) Collection {
	c.Order = append([]string{}, order...)
	return c
}

func (c Collection) WithVariables(rows []VariableRow) Collection {
	c.Variables = append([]VariableRow{}, rows...)
	return c
}

func (c Collection) HasOwnSettings() bool {
	return !c.Auth.IsInherited() || len(c.Headers) > 0 || !c.Scripts.IsEmpty()
}

func (c Collection) WithImportedSettings(imported Collection) Collection {
	c.Auth = imported.Auth
	c.Headers = append([]KeyValue{}, imported.Headers...)
	c.Scripts = imported.Scripts
	return c
}

func (c Collection) ScriptOwner() ScriptOwner {
	if c.IsFolder() {
		return NewScriptOwner(scriptownerkind.Folder, c.ID, c.Name, c.Scripts, c.InheritScripts)
	}
	return NewScriptOwner(scriptownerkind.Collection, c.ID, c.Name, c.Scripts, c.InheritScripts)
}

func (c Collection) OwnVariables() []VariableRow {
	if c.IsFolder() {
		return []VariableRow{}
	}
	return c.Variables
}

func (c Collection) ScopeRef() ScopeRef {
	return NewCollectionScopeRef(c.ID)
}

func (c Collection) MapText(transform func(string) string) Collection {
	c.Auth = c.Auth.MapText(transform)
	c.Headers = mapKeyValues(c.Headers, transform)
	return c
}

// InheritedFields names the auth and header texts a request inherits from this ancestor, as "inherited:<id>.<field>".
func (c Collection) InheritedFields(withAuth bool, headers []int) []TextField {
	prefix := "inherited:" + c.ID
	fields := make([]TextField, 0)
	if withAuth {
		fields = append(fields, c.Auth.TextFields(fieldName(prefix, "auth"))...)
	}
	for _, index := range headers {
		header := c.Headers[index]
		position := strconv.Itoa(index)
		fields = append(fields, NewTextField(fieldName(prefix, "headers", position, "key"), header.Key), NewTextField(fieldName(prefix, "headers", position, "value"), header.Value))
	}
	return fields
}

func (c Collection) InheritableHeaderIndexes(overriding []KeyValue) []int {
	indexes := make([]int, 0)
	for index, header := range c.Headers {
		if header.IsActive() && !isOverridden(header, overriding) {
			indexes = append(indexes, index)
		}
	}
	return indexes
}

func (c Collection) ArrangedChildren(present []string) []string {
	return arrangedNames(c.Order, present)
}

func (c Collection) LogValue() any {
	return map[string]any{"id": c.ID, "kind": c.Kind, "name": c.Name, "description": c.Description, "headers": KeyValues(c.Headers).LogValue()}
}

func RootCollection(ancestors []Collection) (Collection, bool) {
	if len(ancestors) == 0 {
		return Collection{}, false
	}
	return ancestors[0], true
}

func isOverridden(header KeyValue, overriding []KeyValue) bool {
	return slices.ContainsFunc(overriding, func(other KeyValue) bool {
		return other.IsActive() && other.HasKey(header.Key)
	})
}
