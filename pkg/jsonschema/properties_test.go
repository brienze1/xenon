package jsonschema

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

type sampleRow struct {
	Key     string `json:"key" description:"Row key."`
	Enabled *bool  `json:"enabled"`
}

type sampleTarget struct {
	Kind string `json:"kind" enum:"a,b"`
}

type sampleEmbedded struct {
	RequestId string `json:"request_id" description:"Saved request id."`
}

type sample struct {
	sampleEmbedded
	Name     string            `json:"name" description:"The name." required_for:"create,rename" used_by:"create,rename" validated_as:"local_name,x@rename"`
	Count    int               `json:"count"`
	Ratio    float64           `json:"ratio"`
	Flag     bool              `json:"flag" schema_for:"docs_edit"`
	Limit    *int              `json:"limit"`
	Paths    []string          `json:"paths"`
	Rows     []sampleRow       `json:"rows"`
	Target   *sampleTarget     `json:"target"`
	Headers  map[string]string `json:"headers"`
	Depth    json.RawMessage   `json:"depth" schema:"integer|string"`
	Hidden   json.RawMessage   `json:"hidden" schema:"-"`
	Ignored  string            `json:"-"`
	internal string
}

type recursive struct {
	Name     string      `json:"name"`
	Children []recursive `json:"children"`
}

type untypedRaw struct {
	Raw json.RawMessage `json:"raw"`
}

type withInterface struct {
	Any any `json:"any"`
}

func TestPropertiesMapsEveryFieldShape(t *testing.T) {
	properties, err := Properties(reflect.TypeOf(sample{}))
	if err != nil {
		t.Fatalf("Properties() error = %v", err)
	}
	cases := map[string][]string{
		"request_id": {"string"},
		"name":       {"string"},
		"count":      {"integer"},
		"ratio":      {"number"},
		"flag":       {"boolean"},
		"limit":      {"integer"},
		"paths":      {"array"},
		"rows":       {"array"},
		"target":     {"object"},
		"headers":    {"object"},
		"depth":      {"integer", "string"},
	}
	for name, types := range cases {
		property, found := properties[name]
		if !found {
			t.Errorf("property %q missing", name)
			continue
		}
		if !reflect.DeepEqual(property.Types, types) {
			t.Errorf("%s.Types = %v, want %v", name, property.Types, types)
		}
	}
	for _, name := range []string{"hidden", "Ignored", "internal", "-", "sampleEmbedded"} {
		if _, found := properties[name]; found {
			t.Errorf("property %q must not be declared", name)
		}
	}
	if len(properties) != len(cases) {
		t.Errorf("got %d properties, want %d", len(properties), len(cases))
	}
}

func TestPropertiesReadsNestedShapes(t *testing.T) {
	properties, _ := Properties(reflect.TypeOf(sample{}))
	if items := properties["paths"].Items; items == nil || items.Types[0] != "string" {
		t.Errorf("paths.Items = %+v, want string items", items)
	}
	rows := properties["rows"].Items
	if rows == nil || rows.Types[0] != "object" || rows.Properties["key"].Description != "Row key." || rows.Properties["enabled"].Types[0] != "boolean" {
		t.Errorf("rows.Items = %+v", rows)
	}
	if kind := properties["target"].Properties["kind"]; !reflect.DeepEqual(kind.Enum, []string{"a", "b"}) {
		t.Errorf("target.kind.Enum = %v", kind.Enum)
	}
	if values := properties["headers"].AdditionalProperties; values == nil || values.Types[0] != "string" {
		t.Errorf("headers.AdditionalProperties = %+v", values)
	}
}

func TestPropertiesReadsTheArgumentTags(t *testing.T) {
	properties, _ := Properties(reflect.TypeOf(sample{}))
	name := properties["name"]
	if name.Description != "The name." {
		t.Errorf("Description = %q", name.Description)
	}
	if !reflect.DeepEqual(name.RequiredFor, []string{"create", "rename"}) || !reflect.DeepEqual(name.UsedBy, []string{"create", "rename"}) {
		t.Errorf("RequiredFor = %v, UsedBy = %v", name.RequiredFor, name.UsedBy)
	}
	if !reflect.DeepEqual(name.ValidatedAs, []string{"local_name", "x@rename"}) {
		t.Errorf("ValidatedAs = %v", name.ValidatedAs)
	}
	if !reflect.DeepEqual(properties["flag"].SchemaFor, []string{"docs_edit"}) {
		t.Errorf("SchemaFor = %v", properties["flag"].SchemaFor)
	}
	if properties["request_id"].Description != "Saved request id." {
		t.Errorf("embedded field description = %q", properties["request_id"].Description)
	}
}

func TestPropertiesStopsAtARecursiveType(t *testing.T) {
	properties, err := Properties(reflect.TypeOf(recursive{}))
	if err != nil {
		t.Fatalf("Properties() error = %v", err)
	}
	children := properties["children"].Items
	if children == nil || children.Types[0] != "object" {
		t.Fatalf("children.Items = %+v", children)
	}
	if len(children.Properties) != 0 {
		t.Errorf("a type already on the path must become a plain object, got %d properties", len(children.Properties))
	}
}

func TestPropertiesRejectsUntypedFields(t *testing.T) {
	for name, value := range map[string]any{"raw message": untypedRaw{}, "interface": withInterface{}} {
		if _, err := Properties(reflect.TypeOf(value)); err == nil {
			t.Errorf("%s: Properties() error = nil, want an error", name)
		}
	}
	if _, err := Properties(reflect.TypeOf("text")); err == nil {
		t.Error("a non-struct type must be rejected")
	}
}

func TestPropertyMapRendersStandardJsonSchema(t *testing.T) {
	properties, _ := Properties(reflect.TypeOf(sample{}))
	encoded, err := json.Marshal(map[string]any{"name": properties["name"].Map(), "depth": properties["depth"].Map(), "rows": properties["rows"].Map(), "headers": properties["headers"].Map()})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	text := string(encoded)
	for _, want := range []string{
		`"name":{"description":"The name.","type":"string"}`,
		`"depth":{"type":["integer","string"]}`,
		`"items":{"properties":{"enabled":{"type":"boolean"},"key":{"description":"Row key.","type":"string"}},"type":"object"}`,
		`"headers":{"additionalProperties":{"type":"string"},"type":"object"}`,
	} {
		if !strings.Contains(text, want) {
			t.Errorf("rendered schema %s does not contain %s", text, want)
		}
	}
	for _, internal := range []string{"required_for", "RequiredFor", "used_by", "validated_as", "schema_for", "local_name"} {
		if strings.Contains(text, internal) {
			t.Errorf("rendered schema leaks %q: %s", internal, text)
		}
	}
}

func TestPropertyMapRendersEnums(t *testing.T) {
	property := Property{Types: []string{"string"}, Enum: []string{"a", "b"}}
	rendered := property.Map()
	if !reflect.DeepEqual(rendered["enum"], []string{"a", "b"}) || rendered["type"] != "string" {
		t.Errorf("Map() = %v", rendered)
	}
}
