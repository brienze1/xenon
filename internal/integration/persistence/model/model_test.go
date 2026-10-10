package model

import (
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

func write(t *testing.T, fields []yamlnode.Field) string {
	t.Helper()
	document := yamlnode.NewDocument()
	yamlnode.Patch(yamlnode.Root(document), fields)
	out, err := yamlnode.Encode(document)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	return string(out)
}

func read(t *testing.T, source string, out any) {
	t.Helper()
	document, err := yamlnode.Parse([]byte(source))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if err := document.Decode(out); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
}

func contractRequest() entity.Request {
	request := entity.NewRequest("01JA2Z7C2D3E4F5G6H7J8K9M0N", "Create invoice")
	request.Method = "post"
	request.URL = "https://{{host}}/v1/customers/:customerId/invoices?draft=true"
	request.Params = []entity.KeyValue{
		{Key: "draft", Value: "true", Enabled: true},
		{Key: "debug", Value: "1", Description: "only in dev"},
	}
	request.PathVariables = []entity.PathVariable{{Key: "customerId", Value: "42"}}
	request.Headers = []entity.KeyValue{entity.NewKeyValue("Accept", "application/json")}
	request.DisabledAutoHeaders = []string{"User-Agent"}
	request.Body = entity.Body{
		Mode:       bodymode.Raw,
		Raw:        entity.BodyRaw{Language: rawlanguage.Json, Content: "{ \"amount\": 1000 }\n"},
		FormData:   []entity.FormDataRow{},
		UrlEncoded: []entity.KeyValue{},
	}
	return request
}

const contractRequestYaml = `xenon: 1
kind: request
id: 01JA2Z7C2D3E4F5G6H7J8K9M0N
name: Create invoice
description: ""
method: POST
url: https://{{host}}/v1/customers/:customerId/invoices?draft=true
params:
  - key: draft
    value: "true"
    enabled: true
    description: ""
  - key: debug
    value: "1"
    enabled: false
    description: only in dev
pathVariables:
  - key: customerId
    value: "42"
    description: ""
headers:
  - key: Accept
    value: application/json
    enabled: true
    description: ""
autoHeaders:
  disabled:
    - User-Agent
auth:
  type: inherit
body:
  mode: raw
  raw:
    language: json
    content: |
      { "amount": 1000 }
settings:
  timeoutMs: 0
  followRedirects: true
  verifySsl: true
  encodeUrl: true
scripts:
  preRequest: ""
  postResponse: ""
inheritScripts: true
postman: {}
`

func TestRequestIsWrittenInContractOrder(t *testing.T) {
	model := NewRequest(contractRequest())
	if got := write(t, model.Fields()); got != contractRequestYaml {
		t.Fatalf("got:\n%s\nwant:\n%s", got, contractRequestYaml)
	}
}

func TestRequestRoundTrips(t *testing.T) {
	var model Request
	read(t, contractRequestYaml, &model)
	want := contractRequest()
	want.Method = "POST"
	if got := model.ToEntity(); !reflect.DeepEqual(got, want) {
		t.Fatalf("ToEntity() =\n%+v\nwant\n%+v", got, want)
	}
}

func TestMinimalRequestGetsDefaults(t *testing.T) {
	var model Request
	read(t, "xenon: 1\nkind: request\nid: 01R\nname: Ping\n", &model)
	request := model.ToEntity()
	want := entity.NewRequest("01R", "Ping")
	if !reflect.DeepEqual(request, want) {
		t.Fatalf("ToEntity() = %+v, want %+v", request, want)
	}
	stored := NewRequest(request)
	written := write(t, stored.Fields())
	wantYaml := "xenon: 1\nkind: request\nid: 01R\nname: Ping\ndescription: \"\"\nmethod: GET\nurl: \"\"\nparams: []\npathVariables: []\nheaders: []\nautoHeaders:\n  disabled: []\nauth:\n  type: inherit\nbody:\n  mode: none\nsettings:\n  timeoutMs: 0\n  followRedirects: true\n  verifySsl: true\n  encodeUrl: true\nscripts:\n  preRequest: \"\"\n  postResponse: \"\"\ninheritScripts: true\npostman: {}\n"
	if written != wantYaml {
		t.Fatalf("got:\n%s\nwant:\n%s", written, wantYaml)
	}
}

func TestSettingsReadExplicitValues(t *testing.T) {
	var model Settings
	read(t, "timeoutMs: 5000\nfollowRedirects: false\nverifySsl: false\nencodeUrl: false\n", &model)
	want := entity.RequestSettings{TimeoutMs: 5000}
	if got := model.ToEntity(); got != want {
		t.Fatalf("ToEntity() = %+v, want %+v", got, want)
	}
}

func TestAuthKeepsEveryTypedSubObject(t *testing.T) {
	auth := entity.Auth{Type: authtype.Basic, Bearer: entity.AuthBearer{Token: "abc"}, Basic: entity.AuthBasic{Username: "user", Password: "pass"}}
	model := NewAuth(auth)
	want := "auth:\n  type: basic\n  bearer:\n    token: abc\n  basic:\n    username: user\n    password: pass\n"
	if got := write(t, []yamlnode.Field{model.Field()}); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	if got := model.ToEntity(); !reflect.DeepEqual(got, auth) {
		t.Fatalf("ToEntity() = %+v", got)
	}
	bearer := NewAuth(entity.Auth{Type: authtype.Bearer})
	if bearer.Bearer == nil || bearer.Basic != nil {
		t.Fatalf("bearer auth model = %+v", bearer)
	}
	inherit := NewAuth(entity.DefaultAuth())
	if inherit.Bearer != nil || inherit.Basic != nil {
		t.Fatalf("inherit auth model = %+v", inherit)
	}
	var reserved Auth
	read(t, "type: apikey\napikey:\n  key: X\n", &reserved)
	if reserved.ToEntity().Type != authtype.ApiKey {
		t.Fatal("reserved auth types are read")
	}
}

func TestBodyModesAreKeptAcrossSwitches(t *testing.T) {
	body := entity.Body{
		Mode: bodymode.FormData,
		FormData: []entity.FormDataRow{
			{Key: "note", Type: formdatatype.Text, Value: "hello", Enabled: true},
			{Key: "file", Type: formdatatype.File, Src: entity.FileRef{Src: "fixtures/invoice.pdf"}, Enabled: true},
		},
		UrlEncoded: []entity.KeyValue{entity.NewKeyValue("a", "1")},
		Raw:        entity.BodyRaw{Language: rawlanguage.Xml, Content: "<a/>"},
		Binary:     entity.FileRef{Src: "/tmp/payload.bin"},
		GraphQl:    entity.BodyGraphQl{Query: "query { me { id } }", Variables: "{}"},
	}
	model := NewBody(body)
	want := `body:
  mode: formdata
  formdata:
    - key: note
      type: text
      value: hello
      enabled: true
      description: ""
    - key: file
      type: file
      src: fixtures/invoice.pdf
      enabled: true
      description: ""
  urlencoded:
    - key: a
      value: "1"
      enabled: true
      description: ""
  raw:
    language: xml
    content: <a/>
  binary:
    src: /tmp/payload.bin
  graphql:
    query: query { me { id } }
    variables: '{}'
`
	got := write(t, []yamlnode.Field{model.Field()})
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	var decoded struct {
		Body Body `yaml:"body"`
	}
	read(t, got, &decoded)
	if roundTrip := decoded.Body.ToEntity(); !reflect.DeepEqual(roundTrip, body) {
		t.Fatalf("ToEntity() =\n%+v\nwant\n%+v", roundTrip, body)
	}
}

func TestEmptyBodyModesAreOmittedUnlessSelected(t *testing.T) {
	for mode, want := range map[string]string{
		"none":    "body:\n  mode: none\n",
		"raw":     "body:\n  mode: raw\n  raw:\n    language: json\n    content: \"\"\n",
		"binary":  "body:\n  mode: binary\n  binary:\n    src: \"\"\n",
		"graphql": "body:\n  mode: graphql\n  graphql:\n    query: \"\"\n    variables: \"\"\n",
	} {
		body := entity.DefaultBody()
		body.Mode = bodymode.GetBodyMode(mode)
		model := NewBody(body)
		if got := write(t, []yamlnode.Field{model.Field()}); got != want {
			t.Errorf("mode %s got:\n%s\nwant:\n%s", mode, got, want)
		}
	}
	var unknownLanguage Body
	read(t, "mode: raw\nraw:\n  language: yaml\n", &unknownLanguage)
	if unknownLanguage.ToEntity().Raw.Language != rawlanguage.Unknown {
		t.Fatal("unknown raw languages are read as unknown")
	}
}

func TestCollectionAndFolderFiles(t *testing.T) {
	collection := entity.NewCollection("01C", "Billing API")
	collection.Order = []string{"create-customer.yaml", "invoices"}
	collection.Auth = entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: "{{billing_token}}"}}
	collection.Headers = []entity.KeyValue{entity.NewKeyValue("X-Tenant", "acme")}
	model := NewCollection(collection)
	want := `xenon: 1
kind: collection
id: 01C
name: Billing API
description: ""
order:
  - create-customer.yaml
  - invoices
auth:
  type: bearer
  bearer:
    token: '{{billing_token}}'
headers:
  - key: X-Tenant
    value: acme
    enabled: true
    description: ""
variables: []
scripts:
  preRequest: ""
  postResponse: ""
postman: {}
`
	got := write(t, model.Fields())
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	var decoded Collection
	read(t, got, &decoded)
	if roundTrip := decoded.ToEntity(); !reflect.DeepEqual(roundTrip, collection) {
		t.Fatalf("ToEntity() = %+v, want %+v", roundTrip, collection)
	}
	folder := NewCollection(entity.NewFolder("01F", "Invoices"))
	wantFolder := "xenon: 1\nkind: folder\nid: 01F\nname: Invoices\ndescription: \"\"\norder: []\nauth:\n  type: inherit\nheaders: []\nscripts:\n  preRequest: \"\"\n  postResponse: \"\"\ninheritScripts: true\npostman: {}\n"
	if got := write(t, folder.Fields()); got != wantFolder {
		t.Fatalf("got:\n%s\nwant:\n%s", got, wantFolder)
	}
}

func TestWorkspaceFile(t *testing.T) {
	workspace := entity.NewWorkspace("01W", "Acme")
	model := NewWorkspace(workspace)
	want := "xenon: 1\nkind: workspace\nid: 01W\nname: Acme\ndescription: \"\"\norder: []\nvariables: []\nsync: {}\n"
	got := write(t, model.Fields())
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	var decoded Workspace
	read(t, got, &decoded)
	if decoded.Xenon != 1 || decoded.Kind != "workspace" || !reflect.DeepEqual(decoded.ToEntity(), workspace) {
		t.Fatalf("decoded = %+v", decoded)
	}
	var withoutOrder Workspace
	read(t, "name: Acme\n", &withoutOrder)
	if withoutOrder.ToEntity().Order == nil {
		t.Fatal("a missing order reads as empty")
	}
}

func TestItemCopyPatchesTheIdAndOnlyRenamesWhenNamed(t *testing.T) {
	renamed := NewItemCopy(entity.ItemCopy{ID: "copy-1", Name: "Copy"})
	if got, want := write(t, renamed.Fields()), "id: copy-1\nname: Copy\n"; got != want {
		t.Errorf("renamed = %q, want %q", got, want)
	}
	kept := NewItemCopy(entity.ItemCopy{ID: "copy-2"})
	if got, want := write(t, kept.Fields()), "id: copy-2\n"; got != want {
		t.Errorf("kept = %q, want %q", got, want)
	}
}
