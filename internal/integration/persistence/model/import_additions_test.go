package model

import (
	"reflect"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodyencoding"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/importorigin"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

func importedRequest() entity.Request {
	request := entity.NewRequest("01JA2Z7C2D3E4F5G6H7J8K9M0N", "Upload")
	request.Method = "POST"
	request.URL = "{{base}}/upload"
	request.Auth = entity.NewUnsupportedAuth(entity.AuthUnsupported{
		Origin: importorigin.Postman,
		Type:   "hawk",
		Data:   []any{map[string]any{"key": "authId", "value": "id-1", "type": "string"}},
	})
	request.Body = entity.DefaultBody()
	request.Body.Mode = bodymode.FormData
	request.Body.FormData = []entity.FormDataRow{
		{Key: "file", Type: formdatatype.File, Src: entity.NewRepickFileRef("/home/ana/.ssh/id_rsa"), Enabled: true},
		{Key: "note", Type: formdatatype.Text, Value: "hi", Enabled: true},
	}
	request.Body.Binary = entity.NewRepickFileRef("/tmp/payload.bin")
	request.Scripts = entity.Scripts{PreRequest: "pm.environment.set(\"a\", 1);", PostResponse: "pm.test(\"ok\", () => {});\nconsole.log(1);", Disabled: []enums.ScriptPhase{scriptphase.PostResponse}}
	request.Postman = entity.SourceFields{"id": "8c1f", "protocolProfileBehavior": map[string]any{"strictSSL": false}}
	request.Bruno = entity.SourceFields{"seq": 3}
	text := entity.NewSavedExample("01JA2Z7C2D3E4F5G6H7J8K9M0P", "Created")
	text.Request.Method = "POST"
	text.Request.URL = "{{base}}/upload"
	text.Request.Headers = []entity.KeyValue{entity.NewKeyValue("Accept", "application/json")}
	text.Response.Code = 201
	text.Response.Status = "Created"
	text.Response.ContentType = "application/json"
	text.Response.Headers = []entity.KeyValue{entity.NewKeyValue("Content-Type", "application/json")}
	text.Response = text.Response.WithBody([]byte("{\n  \"id\": 42\n}\n"))
	text.Postman = entity.SourceFields{"_postman_previewlanguage": "json", "responseTime": "120"}
	binary := entity.NewSavedExample("01JA2Z7C2D3E4F5G6H7J8K9M0Q", "Image")
	binary.Response = binary.Response.WithBody([]byte{0x89, 0x50, 0x4e, 0x47, 0xff, 0x00})
	request.Examples = []entity.SavedExample{text, binary}
	return request
}

func TestImportedRequestRoundTripsByteStable(t *testing.T) {
	request := importedRequest()
	stored := NewRequest(request)
	first := write(t, stored.Fields())

	var decoded Request
	read(t, first, &decoded)
	roundTrip := decoded.ToEntity()
	if !reflect.DeepEqual(roundTrip, request) {
		t.Fatalf("ToEntity() =\n%+v\nwant\n%+v", roundTrip, request)
	}
	document, err := yamlnode.Parse([]byte(first))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	again := NewRequest(roundTrip)
	yamlnode.Patch(yamlnode.Root(document), again.Fields())
	second, err := yamlnode.Encode(document)
	if err != nil || string(second) != first {
		t.Fatalf("re-save changed the file:\n%s\nwas\n%s", second, first)
	}
	for _, fragment := range []string{
		"srcState: repick",
		"type: unsupported",
		"unsupported:\n    source: postman\n    type: hawk\n    data:\n      - key: authId\n        type: string\n        value: id-1",
		"disabled:\n    - postResponse",
		"bodyEncoding: base64\n      body: iVBOR/8A",
		"bodyEncoding: text\n      body: |\n        {\n          \"id\": 42\n        }",
		"postman:\n  id: 8c1f\n  protocolProfileBehavior:\n    strictSSL: false\nbruno:\n  seq: 3\n",
	} {
		if !strings.Contains(first, fragment) {
			t.Errorf("file misses %q:\n%s", fragment, first)
		}
	}
}

func TestUiSaveKeepsImportOnlyBlocks(t *testing.T) {
	written := NewRequest(importedRequest())
	first := write(t, written.Fields())
	document, err := yamlnode.Parse([]byte(first))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	edited := entity.NewRequest("01JA2Z7C2D3E4F5G6H7J8K9M0N", "Upload")
	edited.Method = "POST"
	edited.URL = "{{base}}/upload"
	edited.Auth = entity.Auth{Type: authtype.Unsupported}
	edited.Body = importedRequest().Body
	edited.Body.FormData[0].Src = entity.FileRef{Src: "/home/ana/new.txt"}
	edited.Body.Binary = entity.FileRef{Src: "/tmp/new.bin"}
	stored := NewRequest(edited)
	yamlnode.Patch(yamlnode.Root(document), stored.Fields())
	out, err := yamlnode.Encode(document)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	var decoded Request
	read(t, string(out), &decoded)
	saved := decoded.ToEntity()
	original := importedRequest()
	if !reflect.DeepEqual(saved.Scripts, original.Scripts) || !reflect.DeepEqual(saved.Examples, original.Examples) || !reflect.DeepEqual(saved.Postman, original.Postman) || !reflect.DeepEqual(saved.Auth.Unsupported, original.Auth.Unsupported) {
		t.Fatalf("import-only blocks were not kept:\n%s", out)
	}
	if saved.Body.FormData[0].Src.NeedsRepick() || saved.Body.Binary.NeedsRepick() {
		t.Fatalf("a re-picked file must lose its repick state:\n%s", out)
	}
}

func TestReservedAuthAttributesFollowTheType(t *testing.T) {
	auth := entity.Auth{Type: authtype.ApiKey, Attributes: []entity.AuthAttribute{{Key: "key", Value: "X-Key"}, {Key: "value", Value: "{{k}}"}, {Key: "in", Value: "header"}}}
	stored := NewAuth(auth)
	got := write(t, []yamlnode.Field{stored.Field()})
	want := "auth:\n  type: apikey\n  apikey:\n    key: X-Key\n    value: '{{k}}'\n    in: header\n"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	var decoded struct {
		Auth Auth `yaml:"auth"`
	}
	read(t, got, &decoded)
	if roundTrip := decoded.Auth.ToEntity(); !reflect.DeepEqual(roundTrip, auth) {
		t.Fatalf("ToEntity() = %+v", roundTrip)
	}
	for _, authType := range []string{"oauth2", "awsv4", "digest"} {
		var other struct {
			Auth Auth `yaml:"auth"`
		}
		read(t, "auth:\n  type: "+authType+"\n  "+authType+":\n    a: 1\n  apikey:\n    key: kept\n", &other)
		attributes := other.Auth.ToEntity().Attributes
		if len(attributes) != 1 || attributes[0].Key != "a" || attributes[0].Value != 1 {
			t.Errorf("%s attributes = %+v", authType, attributes)
		}
	}
	var broken struct {
		Auth Auth `yaml:"auth"`
	}
	read(t, "auth:\n  type: digest\n  digest: plain\n", &broken)
	if attributes := broken.Auth.ToEntity().Attributes; attributes != nil {
		t.Fatalf("a non-mapping sub-key has no attributes, got %+v", attributes)
	}
}

func TestScriptsDropUnknownDisabledPhases(t *testing.T) {
	var stored Scripts
	read(t, "preRequest: a\npostResponse: b\ndisabled: [preRequest, weird]\n", &stored)
	if got := stored.ToEntity(); !reflect.DeepEqual(got, entity.Scripts{PreRequest: "a", PostResponse: "b", Disabled: []enums.ScriptPhase{scriptphase.PreRequest}}) {
		t.Fatalf("ToEntity() = %+v", got)
	}
	empty := NewScripts(entity.Scripts{})
	if got := write(t, []yamlnode.Field{{Key: "scripts", Value: empty.Node()}}); got != "scripts:\n  preRequest: \"\"\n  postResponse: \"\"\n" {
		t.Fatalf("empty scripts = %q", got)
	}
}

func TestContainerScriptsAndSourceFields(t *testing.T) {
	collection := entity.NewCollection("01C", "Billing")
	collection.Scripts = entity.Scripts{PreRequest: "pm.globals.set('a', 1);"}
	collection.Postman = entity.SourceFields{"_postman_id": "abc", "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"}
	collection.Variables = []entity.VariableRow{{Key: "page", InitialValue: "1", Enabled: true, Postman: entity.SourceFields{"type": "string"}}}
	stored := NewCollection(collection)
	got := write(t, stored.Fields())
	for _, fragment := range []string{"scripts:\n  preRequest: pm.globals.set('a', 1);\n  postResponse: \"\"\n", "postman:\n  _postman_id: abc", "    description: \"\"\n    postman:\n      type: string\n"} {
		if !strings.Contains(got, fragment) {
			t.Errorf("collection file misses %q:\n%s", fragment, got)
		}
	}
	var decoded Collection
	read(t, got, &decoded)
	if roundTrip := decoded.ToEntity(); !reflect.DeepEqual(roundTrip, collection) {
		t.Fatalf("ToEntity() = %+v\nwant %+v", roundTrip, collection)
	}
	folder := entity.NewFolder("01F", "Invoices")
	folder.Bruno = entity.SourceFields{"seq": 2}
	storedFolder := NewCollection(folder)
	folderFile := write(t, storedFolder.Fields())
	if !strings.Contains(folderFile, "inheritScripts: true\npostman: {}\nbruno:\n  seq: 2\n") {
		t.Fatalf("folder file:\n%s", folderFile)
	}
}

func TestEnvironmentAndWorkspaceSourceFields(t *testing.T) {
	environment := entity.NewEnvironment("01E", "dev")
	environment.Postman = entity.SourceFields{"_postman_variable_scope": "environment"}
	environment.Bruno = entity.SourceFields{"color": "#ff0000"}
	storedEnvironment := NewEnvironment(environment)
	got := write(t, storedEnvironment.Fields())
	if !strings.Contains(got, "variables: []\npostman:\n  _postman_variable_scope: environment\nbruno:\n  color: '#ff0000'\n") {
		t.Fatalf("environment file:\n%s", got)
	}
	var decodedEnvironment Environment
	read(t, got, &decodedEnvironment)
	if roundTrip := decodedEnvironment.ToEntity(""); !reflect.DeepEqual(roundTrip, environment) {
		t.Fatalf("environment ToEntity() = %+v", roundTrip)
	}
	plainEnvironment := NewEnvironment(entity.NewEnvironment("01E", "qa"))
	plain := write(t, plainEnvironment.Fields())
	if strings.Contains(plain, "postman") || strings.Contains(plain, "bruno") {
		t.Fatalf("an environment without passthrough writes none:\n%s", plain)
	}
	workspace := entity.NewWorkspace("01W", "Acme")
	workspace.Postman = entity.SourceFields{"_postman_exported_using": "Postman/11"}
	storedWorkspace := NewWorkspace(workspace)
	workspaceFile := write(t, storedWorkspace.Fields())
	if !strings.Contains(workspaceFile, "sync: {}\npostman:\n  _postman_exported_using: Postman/11\n") {
		t.Fatalf("workspace file:\n%s", workspaceFile)
	}
	var decodedWorkspace Workspace
	read(t, workspaceFile, &decodedWorkspace)
	if roundTrip := decodedWorkspace.ToEntity(); !reflect.DeepEqual(roundTrip, workspace) {
		t.Fatalf("workspace ToEntity() = %+v", roundTrip)
	}
}

func TestExampleResponseDefaultsAndRequest(t *testing.T) {
	var stored Example
	read(t, "id: x\nname: Empty\nrequest:\n  url: /a\nresponse:\n  code: 404\n  body: nope\n", &stored)
	example := stored.ToEntity()
	if example.Request.Method != "GET" || example.Response.Code != 404 || string(example.Response.Body) != "nope" || example.Response.BodyEncoding != bodyencoding.Text || example.Postman != nil {
		t.Fatalf("ToEntity() = %+v", example)
	}
	if examplesToEntity(nil) != nil {
		t.Fatal("no examples stay nil")
	}
	request := entity.NewRequest("01R", "R")
	plainRequest := NewRequest(request)
	if got := write(t, plainRequest.Fields()); strings.Contains(got, "examples") || strings.Contains(got, "bruno") || strings.Contains(got, "srcState") {
		t.Fatalf("a plain request writes no import blocks:\n%s", got)
	}
}
