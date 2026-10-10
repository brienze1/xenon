package model

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/envcolor"
)

func contractVariables() []entity.VariableRow {
	return []entity.VariableRow{
		{Key: "base_url", InitialValue: "https://api.acme.com", Enabled: true},
		{Key: "api_key", Secret: true, Enabled: true, Description: "from the vendor portal"},
		{Key: "legacy", InitialValue: "x", Enabled: false},
	}
}

func TestEnvironmentFileFollowsTheContract(t *testing.T) {
	environment := entity.NewEnvironment("01JA3B4C5D6E7F8G9H0J1K2M3N", "prod")
	environment.Variables = contractVariables()
	stored := NewEnvironment(environment)
	want := `xenon: 1
kind: environment
id: 01JA3B4C5D6E7F8G9H0J1K2M3N
name: prod
description: ""
color: red
confirmBeforeSend: true
variables:
  - key: base_url
    value: https://api.acme.com
    enabled: true
    description: ""
  - key: api_key
    secret: true
    enabled: true
    description: from the vendor portal
  - key: legacy
    value: x
    enabled: false
    description: ""
`
	got := write(t, stored.Fields())
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	var decoded Environment
	read(t, got, &decoded)
	roundTrip := decoded.ToEntity("prod.yaml")
	environment.FileName = "prod.yaml"
	if !reflect.DeepEqual(roundTrip, environment) {
		t.Fatalf("ToEntity() =\n%+v\nwant\n%+v", roundTrip, environment)
	}
}

func TestEnvironmentSecretRowsNeverWriteAValue(t *testing.T) {
	environment := entity.NewEnvironment("01E", "dev")
	environment.Variables = []entity.VariableRow{{Key: "api_key", InitialValue: "leaked", Secret: true, Enabled: true}}
	stored := NewEnvironment(environment)
	want := "xenon: 1\nkind: environment\nid: 01E\nname: dev\ndescription: \"\"\ncolor: none\nconfirmBeforeSend: false\nvariables:\n  - key: api_key\n    secret: true\n    enabled: true\n    description: \"\"\n"
	if got := write(t, stored.Fields()); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	var hand Environment
	read(t, "id: 01E\nname: dev\nvariables:\n  - key: api_key\n    secret: true\n    value: leaked\n", &hand)
	row := hand.ToEntity("dev.yaml").Variables[0]
	if !row.Secret || !row.HasStraySecretValue() || !row.Enabled {
		t.Fatalf("a hand-written secret value is read as stray: %+v", row)
	}
}

func TestEnvironmentColorsAndDefaults(t *testing.T) {
	cases := map[string]any{"": envcolor.None, "cyan": envcolor.Cyan, "MAGENTA": envcolor.Magenta, "purple": envcolor.None}
	for raw, want := range cases {
		stored := Environment{Color: raw}
		if got := stored.ToEntity("e.yaml").Color; got != want {
			t.Errorf("color %q = %q, want %q", raw, got, want)
		}
	}
	var minimal Environment
	read(t, "id: 01E\nname: local\n", &minimal)
	environment := minimal.ToEntity("local.yaml")
	if environment.Variables == nil || len(environment.Variables) != 0 || environment.ConfirmBeforeSend || environment.FileName != "local.yaml" {
		t.Errorf("minimal environment = %+v", environment)
	}
}

func TestCollectionAndWorkspaceVariablesRoundTrip(t *testing.T) {
	collection := entity.NewCollection("01C", "Billing")
	collection.Variables = contractVariables()
	stored := NewCollection(collection)
	got := write(t, stored.Fields())
	want := "variables:\n  - key: base_url\n    value: https://api.acme.com\n    enabled: true\n    description: \"\"\n  - key: api_key\n    secret: true\n"
	if !strings.Contains(got, want) {
		t.Fatalf("collection.yaml:\n%s\nmissing:\n%s", got, want)
	}
	var decoded Collection
	read(t, got, &decoded)
	if roundTrip := decoded.ToEntity(); !reflect.DeepEqual(roundTrip.Variables, collection.Variables) {
		t.Fatalf("collection variables = %+v", roundTrip.Variables)
	}
	workspace := entity.NewWorkspace("01W", "Acme")
	workspace.Variables = contractVariables()
	storedWorkspace := NewWorkspace(workspace)
	var decodedWorkspace Workspace
	read(t, write(t, storedWorkspace.Fields()), &decodedWorkspace)
	if roundTrip := decodedWorkspace.ToEntity(); !reflect.DeepEqual(roundTrip, workspace) {
		t.Fatalf("workspace = %+v", roundTrip)
	}
	folder := entity.NewFolder("01F", "Invoices")
	folder.Variables = contractVariables()
	storedFolder := NewCollection(folder)
	if got := write(t, storedFolder.Fields()); strings.Contains(got, "variables") {
		t.Fatalf("folder.yaml never gets variables:\n%s", got)
	}
}

func TestLocalVariablesJson(t *testing.T) {
	local := entity.NewLocalVariables().WithActiveEnvironment("01E")
	local.Current["env:01E"] = map[string]string{"base_url": "http://localhost:8080"}
	local.Current["globals"] = map[string]string{}
	data, err := json.Marshal(NewLocalVariables(local))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"current":{"env:01E":{"base_url":"http://localhost:8080"}},"activeEnvironmentId":"01E","version":1}`
	if string(data) != want {
		t.Fatalf("variables.json = %s, want %s", data, want)
	}
	var decoded LocalVariables
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	roundTrip := decoded.ToEntity()
	if roundTrip.ActiveEnvironmentID != "01E" || roundTrip.Current["env:01E"]["base_url"] != "http://localhost:8080" || len(roundTrip.Current) != 1 {
		t.Fatalf("ToEntity() = %+v", roundTrip)
	}
	var empty LocalVariables
	if fresh := empty.ToEntity(); fresh.Current == nil || fresh.HasActiveEnvironment() {
		t.Fatalf("an empty file reads as no local state: %+v", fresh)
	}
}
