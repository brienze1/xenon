package dto

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
)

type testEditArguments struct {
	Id   string `json:"id" description:"Item id." used_by:"switch,rename" required_for:"switch,rename" validated_as:"local_id"`
	Name string `json:"name" description:"Item name." used_by:"create,rename" required_for:"create,rename"`
	Path string `json:"path" description:"Item path." used_by:"create,open" required_for:"create,open"`
}

type testMoveArguments struct {
	Id    string `json:"id" description:"Moved item id." required_for:"move"`
	Index int    `json:"index" description:"New position." required_for:"always"`
}

type testConflictArguments struct {
	Id int `json:"id" description:"A numeric id."`
}

type testSharedArguments struct {
	CollectionId string `json:"collection_id" description:"Collection id." required_for:"get,plan,apply" validated_as:"id"`
	Yaml         string `json:"yaml" description:"The document." used_by:"plan,apply"`
	Confirm      bool   `json:"confirm_deletes" description:"Confirm deletions." schema_for:"docs_edit"`
}

type testTarget struct {
	Kind string `json:"kind" description:"Target kind." validated_as:"item_kind"`
}

type testResolution struct {
	Id string `json:"id" validated_as:"item_id"`
}

type testNestedArguments struct {
	Target  testTarget       `json:"target" description:"What to act on."`
	Items   []testResolution `json:"items" description:"Resolutions."`
	OnClash string           `json:"on_clash" description:"Clash rule." validated_as:"apply_to_all,items[].resolution,local_id@workspace"`
}

func testTool(name string, actions []string, arguments ...MCPToolArguments) MCPTool {
	return NewMCPTool(entity.AITool{Name: name, Description: name, Actions: actions, Effect: aitooleffect.Write}, arguments...)
}

func schemaProperty(t *testing.T, schema map[string]any, name string) map[string]any {
	t.Helper()
	properties, _ := schema["properties"].(map[string]any)
	property, found := properties[name].(map[string]any)
	if !found {
		t.Fatalf("property %q missing from %v", name, schema)
	}
	return property
}

func TestInputSchemaDeclaresEveryArgumentWithActionsInDescriptions(t *testing.T) {
	tool := testTool("workspace_edit", []string{"create", "open", "switch", "rename"}, NewMCPToolArguments[testEditArguments]("create", "open", "switch", "rename"))
	schema := tool.InputSchema()
	cases := map[string]string{
		"id":   "Item id. Used by switch and rename. Required for switch and rename.",
		"name": "Item name. Used by create and rename. Required for create and rename.",
		"path": "Item path. Used by create and open. Required for create and open.",
	}
	for name, description := range cases {
		property := schemaProperty(t, schema, name)
		if property["type"] != "string" || property["description"] != description {
			t.Errorf("%s = %v, want description %q", name, property, description)
		}
	}
	action := schemaProperty(t, schema, "action")
	if !reflect.DeepEqual(action["enum"], []string{"create", "open", "switch", "rename"}) {
		t.Errorf("action = %v", action)
	}
	if !reflect.DeepEqual(schema["required"], []string{"action"}) {
		t.Errorf("required = %v", schema["required"])
	}
}

func TestInputSchemaMergesRegistrationsSharingAField(t *testing.T) {
	tool := testTool("env_edit", []string{"create", "open", "switch", "rename", "move"},
		NewMCPToolArguments[testEditArguments]("create", "open", "switch", "rename"),
		NewMCPToolArguments[testMoveArguments]("move"))
	schema := tool.InputSchema()
	id := schemaProperty(t, schema, "id")
	if id["description"] != "Item id. Used by switch, rename and move. Required for switch, rename and move." {
		t.Errorf("id = %v", id)
	}
	if index := schemaProperty(t, schema, "index"); index["type"] != "integer" || index["description"] != "New position. Used by move." {
		t.Errorf("index = %v", index)
	}
	if !reflect.DeepEqual(schema["required"], []string{"action", "index"}) {
		t.Errorf("required = %v", schema["required"])
	}
}

func TestArgumentPropertiesReportsATypeConflict(t *testing.T) {
	tool := testTool("tool", []string{"create", "open", "switch", "rename", "get"},
		NewMCPToolArguments[testEditArguments]("create", "open", "switch", "rename"),
		NewMCPToolArguments[testConflictArguments]("get"))
	if _, err := tool.ArgumentProperties(); err == nil || !strings.Contains(err.Error(), "id") {
		t.Errorf("ArgumentProperties() error = %v, want a conflict on id", err)
	}
	if schemaProperty(t, tool.InputSchema(), "id")["type"] != "string" {
		t.Error("the first registration must win at runtime")
	}
}

func TestInputSchemaScopesSharedFieldsToTheirTool(t *testing.T) {
	docs := testTool("docs", []string{"get", "plan"}, NewMCPToolArguments[testSharedArguments]("get", "plan")).InputSchema()
	edit := testTool("docs_edit", []string{"apply"}, NewMCPToolArguments[testSharedArguments]("apply")).InputSchema()
	if schemaProperty(t, docs, "yaml")["description"] != "The document. Used by plan." {
		t.Errorf("docs yaml = %v", schemaProperty(t, docs, "yaml"))
	}
	if schemaProperty(t, docs, "collection_id")["description"] != "Collection id. Used by get and plan. Required for get and plan." {
		t.Errorf("docs collection_id = %v", schemaProperty(t, docs, "collection_id"))
	}
	if properties := docs["properties"].(map[string]any); properties["confirm_deletes"] != nil {
		t.Error("docs must not declare a field scoped to docs_edit")
	}
	if schemaProperty(t, edit, "confirm_deletes")["description"] != "Confirm deletions." {
		t.Errorf("a single-action tool gets no Used by: %v", schemaProperty(t, edit, "confirm_deletes"))
	}
	if schemaProperty(t, edit, "collection_id")["description"] != "Collection id. Required for apply." {
		t.Errorf("docs_edit collection_id = %v", schemaProperty(t, edit, "collection_id"))
	}
}

func TestInputSchemaOfAToolWithoutActions(t *testing.T) {
	schema := testTool("send", nil, NewMCPToolArguments[testMoveArguments]()).InputSchema()
	properties := schema["properties"].(map[string]any)
	if properties["action"] != nil {
		t.Error("a tool without actions has no selector")
	}
	if schemaProperty(t, schema, "id")["description"] != "Moved item id." {
		t.Errorf("id = %v", schemaProperty(t, schema, "id"))
	}
	if !reflect.DeepEqual(schema["required"], []string{"index"}) {
		t.Errorf("required = %v", schema["required"])
	}
}

func TestInputSchemaHasNoCombinatorsAndRendersNestedShapes(t *testing.T) {
	schema := testTool("export", []string{"run"}, NewMCPToolArguments[testNestedArguments]("run")).InputSchema()
	encoded, _ := json.Marshal(schema)
	for _, keyword := range []string{"oneOf", "anyOf", "allOf", "validated_as", "item_kind"} {
		if strings.Contains(string(encoded), keyword) {
			t.Errorf("schema contains %q: %s", keyword, encoded)
		}
	}
	target := schemaProperty(t, schema, "target")
	if target["type"] != "object" || target["properties"].(map[string]any)["kind"] == nil {
		t.Errorf("target = %v", target)
	}
}

func TestArgumentOfMapsValidationAttributesBackToArguments(t *testing.T) {
	tool := testTool("tool", []string{"run", "workspace"}, NewMCPToolArguments[testNestedArguments]("run", "workspace"), NewMCPToolArguments[testEditArguments]("run"))
	cases := []struct{ action, attribute, want string }{
		{"run", "item_kind", "target.kind"},
		{"run", "item_id", "items[].id"},
		{"run", "apply_to_all", "on_clash"},
		{"run", "items[3].resolution", "on_clash"},
		{"run", "local_id", "id"},
		{"workspace", "local_id", "on_clash"},
		{"run", "tab_id", "tab_id"},
		{"run", "path", "path"},
		{"run", "item_kinds", "item_kinds"},
	}
	for _, tc := range cases {
		if got := tool.ArgumentOf(tc.action, tc.attribute); got != tc.want {
			t.Errorf("ArgumentOf(%q, %q) = %q, want %q", tc.action, tc.attribute, got, tc.want)
		}
	}
}
