package jsonschema

// Property is the JSON Schema of one field, plus the argument metadata read from its struct tags.
type Property struct {
	Types                []string
	Items                *Property
	Properties           map[string]Property
	AdditionalProperties *Property
	Enum                 []string
	Description          string
	RequiredFor          []string
	UsedBy               []string
	ValidatedAs          []string
	SchemaFor            []string
}

// Map renders the standard JSON Schema keywords; the argument metadata is never rendered.
func (p Property) Map() map[string]any {
	schema := map[string]any{"type": p.renderedType()}
	if p.Items != nil {
		schema["items"] = p.Items.Map()
	}
	if len(p.Properties) > 0 {
		properties := make(map[string]any, len(p.Properties))
		for name, property := range p.Properties {
			properties[name] = property.Map()
		}
		schema["properties"] = properties
	}
	if p.AdditionalProperties != nil {
		schema["additionalProperties"] = p.AdditionalProperties.Map()
	}
	if len(p.Enum) > 0 {
		schema["enum"] = p.Enum
	}
	if p.Description != "" {
		schema["description"] = p.Description
	}
	return schema
}

func (p Property) WithDescription(description string) Property {
	p.Description = description
	return p
}

func (p Property) renderedType() any {
	if len(p.Types) == 1 {
		return p.Types[0]
	}
	return p.Types
}
