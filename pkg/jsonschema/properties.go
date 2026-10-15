package jsonschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

const (
	jsonTag        = "json"
	schemaTag      = "schema"
	descriptionTag = "description"
	enumTag        = "enum"
	requiredForTag = "required_for"
	usedByTag      = "used_by"
	validatedAsTag = "validated_as"
	schemaForTag   = "schema_for"
	omittedName    = "-"
	listSeparator  = ","
	typeSeparator  = "|"
	stringType     = "string"
	integerType    = "integer"
	numberType     = "number"
	booleanType    = "boolean"
	arrayType      = "array"
	objectType     = "object"
)

var rawMessageType = reflect.TypeOf(json.RawMessage{})

// Properties reflects the exported fields of a struct type into JSON Schema properties keyed by their json names.
// Anonymous embedded structs are flattened, as encoding/json decodes them.
func Properties(t reflect.Type) (map[string]Property, error) {
	t = dereference(t)
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s is not a struct", t)
	}
	return structProperties(t, map[reflect.Type]bool{t: true})
}

func structProperties(t reflect.Type, path map[reflect.Type]bool) (map[string]Property, error) {
	properties := map[string]Property{}
	for index := range t.NumField() {
		field := t.Field(index)
		if !field.IsExported() && !field.Anonymous {
			continue
		}
		name, named := jsonName(field)
		if name == omittedName || field.Tag.Get(schemaTag) == omittedName {
			continue
		}
		if field.Anonymous && !named && dereference(field.Type).Kind() == reflect.Struct {
			embedded, err := structProperties(dereference(field.Type), path)
			if err != nil {
				return nil, err
			}
			for embeddedName, property := range embedded {
				properties[embeddedName] = property
			}
			continue
		}
		if !field.IsExported() {
			continue
		}
		property, err := fieldProperty(field, path)
		if err != nil {
			return nil, err
		}
		properties[name] = property
	}
	return properties, nil
}

func fieldProperty(field reflect.StructField, path map[reflect.Type]bool) (Property, error) {
	property, err := typeProperty(field.Type, field.Tag.Get(schemaTag), path)
	if err != nil {
		return Property{}, fmt.Errorf("field %s: %w", field.Name, err)
	}
	property.Description = field.Tag.Get(descriptionTag)
	property.Enum = list(field.Tag.Get(enumTag))
	property.RequiredFor = list(field.Tag.Get(requiredForTag))
	property.UsedBy = list(field.Tag.Get(usedByTag))
	property.ValidatedAs = list(field.Tag.Get(validatedAsTag))
	property.SchemaFor = list(field.Tag.Get(schemaForTag))
	return property, nil
}

func typeProperty(t reflect.Type, override string, path map[reflect.Type]bool) (Property, error) {
	if override != "" {
		return Property{Types: strings.Split(override, typeSeparator)}, nil
	}
	if t == rawMessageType {
		return Property{}, fmt.Errorf("a json.RawMessage needs a schema tag")
	}
	t = dereference(t)
	switch t.Kind() {
	case reflect.String:
		return Property{Types: []string{stringType}}, nil
	case reflect.Bool:
		return Property{Types: []string{booleanType}}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Property{Types: []string{integerType}}, nil
	case reflect.Float32, reflect.Float64:
		return Property{Types: []string{numberType}}, nil
	case reflect.Slice, reflect.Array:
		items, err := typeProperty(t.Elem(), "", path)
		if err != nil {
			return Property{}, err
		}
		return Property{Types: []string{arrayType}, Items: &items}, nil
	case reflect.Map:
		values, err := typeProperty(t.Elem(), "", path)
		if err != nil {
			return Property{}, err
		}
		return Property{Types: []string{objectType}, AdditionalProperties: &values}, nil
	case reflect.Struct:
		return objectProperty(t, path)
	default:
		return Property{}, fmt.Errorf("%s has no JSON Schema type; add a schema tag", t)
	}
}

func objectProperty(t reflect.Type, path map[reflect.Type]bool) (Property, error) {
	if path[t] {
		return Property{Types: []string{objectType}}, nil
	}
	path[t] = true
	defer delete(path, t)
	properties, err := structProperties(t, path)
	if err != nil {
		return Property{}, err
	}
	return Property{Types: []string{objectType}, Properties: properties}, nil
}

func jsonName(field reflect.StructField) (string, bool) {
	name := strings.SplitN(field.Tag.Get(jsonTag), listSeparator, 2)[0]
	if name == "" {
		return field.Name, false
	}
	return name, true
}

func dereference(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func list(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, listSeparator)
}
