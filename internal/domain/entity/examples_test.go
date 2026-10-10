package entity

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/bodyencoding"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func TestExampleResponseCodeMustBeAnHttpStatus(t *testing.T) {
	for code, valid := range map[int]bool{99: false, 100: true, 599: true, 600: false} {
		if got := (ExampleResponse{Code: code}).HasValidCode(); got != valid {
			t.Errorf("HasValidCode(%d) = %v", code, got)
		}
	}
}

func TestExampleBodiesUpTo1MBOfTextAreInline(t *testing.T) {
	text := NewExampleResponse().WithBody(bytes.Repeat([]byte("a"), InlineBodyLimit))
	if !text.HasInlineBody() || text.WithBody(bytes.Repeat([]byte("a"), InlineBodyLimit+1)).HasInlineBody() {
		t.Error("only text bodies up to the inline limit are inline")
	}
	if binary := NewExampleResponse().WithBinaryBody([]byte("a")); binary.HasInlineBody() || !binary.IsBinary() {
		t.Error("binary bodies are never inline")
	}
	if invalid := NewExampleResponse().WithBody([]byte{0xff}); invalid.BodyEncoding != bodyencoding.Base64 {
		t.Error("bodies that are not UTF-8 are stored as base64")
	}
}

func TestExampleFromExchangeCopiesTheSentRequestAndTheResponse(t *testing.T) {
	sent := NewRequest("r1", "Create")
	sent.Method = "post"
	sent.URL = "{{base}}/x"
	sent.Headers = []KeyValue{NewKeyValue("A", "1")}
	response := NewResponse(201, "Created", "HTTP/1.1", []KeyValue{NewKeyValue("Content-Type", "image/png")}, nil, 0, Timings{})
	response.Binary = true
	example := NewExampleFromExchange("e1", " Pixel ", sent, response, []byte("png"))
	if example.Name != "Pixel" || example.Request.Method != "POST" || example.Request.URL != "{{base}}/x" || len(example.Request.Headers) != 1 {
		t.Errorf("request snapshot = %+v", example.Request)
	}
	if example.Response.Code != 201 || example.Response.Status != "Created" || example.Response.ContentType != "image/png" || !example.Response.IsBinary() {
		t.Errorf("response = %+v", example.Response)
	}
	sent.Headers[0].Value = "changed"
	if example.Request.Headers[0].Value != "1" {
		t.Error("the snapshot does not share rows with the request")
	}
	served := example.Response.AsResponse()
	if served.ContentType != "image/png" || !served.Binary || string(served.Body) != "png" || served.StatusCode != 201 {
		t.Errorf("AsResponse() = %+v", served)
	}
}

func TestExampleEditsKeepTheirPassthroughAndOptionallyTheStoredBody(t *testing.T) {
	stored := NewSavedExample("e1", "Created")
	stored.Postman = SourceFields{"id": "pm"}
	stored.Response = stored.Response.WithBinaryBody([]byte{1, 2})
	edit := NewSavedExample("e1", " Renamed ")
	edit.Response = edit.Response.WithBody([]byte("new"))
	kept := stored.Edited(edit, true)
	if kept.Name != "Renamed" || kept.Postman["id"] != "pm" || !kept.Response.IsBinary() || !bytes.Equal(kept.Response.Body, []byte{1, 2}) {
		t.Errorf("Edited(keep) = %+v", kept)
	}
	if replaced := stored.Edited(edit, false); string(replaced.Response.Body) != "new" || replaced.Response.IsBinary() {
		t.Errorf("Edited(replace) = %+v", replaced.Response)
	}
}

func TestRequestExamplesAreEditedInPlace(t *testing.T) {
	first, second := NewSavedExample("e1", "One"), NewSavedExample("e2", "Two")
	request := NewRequest("r1", "R").WithExampleAppended(first).WithExampleAppended(second)
	copied := first.Duplicated("e3")
	if copied.Name != "One Copy" || copied.ID != "e3" {
		t.Errorf("Duplicated() = %+v", copied)
	}
	duplicated := request.WithExampleAfter("e1", copied)
	if ids := exampleIDs(duplicated); ids != "e1,e3,e2" {
		t.Errorf("WithExampleAfter() = %s", ids)
	}
	if ids := exampleIDs(duplicated.WithoutExample("e3")); ids != "e1,e2" || exampleIDs(duplicated) != "e1,e3,e2" {
		t.Errorf("WithoutExample() = %s", ids)
	}
	renamed := request.WithExample(second.Renamed(" Deux "))
	if found, ok := renamed.Example("e2"); !ok || found.Name != "Deux" {
		t.Errorf("WithExample() = %+v", found)
	}
	if exampleIDs(request.WithExample(NewSavedExample("x", "X"))) != "e1,e2" {
		t.Error("an unknown example is not added")
	}
	if _, ok := request.Example("x"); ok {
		t.Error("Example(unknown) is found")
	}
	details := NewExampleDetails(Request{ID: "r1", Revision: "rev"}, first)
	if details.RequestID != "r1" || details.Revision != "rev" || details.ServesBodyByRoute() || first.BodyKey() != "example:e1" {
		t.Errorf("details = %+v", details)
	}
}

func TestTabKindFollowsTheExample(t *testing.T) {
	if NewTabKind("") != itemkind.Request || NewTabKind("e1") != itemkind.Example {
		t.Error("an example id makes an example tab")
	}
}

func TestSentRequestOfTheStoredResponse(t *testing.T) {
	fallback := NewRequest("r1", "Saved")
	response := NewResponse(200, "OK", "", nil, nil, 0, Timings{})
	if got := response.SentRequestOr(fallback); got.Name != "Saved" {
		t.Error("without a sent request the fallback is used")
	}
	if got := response.WithSentRequest(NewRequest("r1", "Sent")).SentRequestOr(fallback); got.Name != "Sent" {
		t.Error("the sent request wins")
	}
}

func exampleIDs(request Request) string {
	ids := make([]string, 0, len(request.Examples))
	for _, example := range request.Examples {
		ids = append(ids, example.ID)
	}
	return strings.Join(ids, ",")
}

func TestImportedVariablesItem(t *testing.T) {
	environment := NewEnvironment("env1", "dev")
	environment.Variables = []VariableRow{NewVariableRow("a", "1"), {Key: "s", Secret: true, Enabled: true}}
	variables := NewImportVariables(environment, map[string]string{"s": "v", "a": "ignored"})
	item := NewVariablesImportItem("3", "environment", variables)
	if item.Name != "dev" || !item.IsVariables() || !reflect.DeepEqual(item.Counts(), ImportCounts{Environments: 1, Variables: 2}) || len(item.Scripts()) != 0 {
		t.Errorf("item = %+v counts %+v", item, item.Counts())
	}
	if value, found := variables.SecretOf(environment.Variables[1]); !found || value != "v" {
		t.Error("the secret value is found for a secret row")
	}
	if _, found := variables.SecretOf(environment.Variables[0]); found {
		t.Error("plain rows have no secret value")
	}
	if globals := NewVariablesImportItem("4", "globals", variables); !reflect.DeepEqual(globals.Counts(), ImportCounts{Variables: 2}) {
		t.Errorf("globals counts = %+v", globals.Counts())
	}
	if logged := variables.LogValue(); !reflect.DeepEqual(logged, map[string]any{"name": "dev", "variables": 2}) {
		t.Errorf("LogValue() = %v", logged)
	}
}

func TestVariablesOutcome(t *testing.T) {
	environment := NewEnvironment("env1", "dev")
	environment.Variables = []VariableRow{NewVariableRow("a", "1")}
	item := NewVariablesImportItem("0", "environment", NewImportVariables(environment, nil))
	replaced := environment
	outcome := item.CompletedVariables(ImportVariablesWrite{Environment: &environment, Replaced: &replaced}, "L")
	if outcome.Item.Status != "replaced" || outcome.Item.ID != "env1" || outcome.Item.Name != "dev" || !outcome.Replaced || outcome.Created.Environments != 1 {
		t.Errorf("outcome = %+v", outcome)
	}
	workspace := NewWorkspace("w1", "Acme")
	kept := []VariableRow{NewVariableRow("region", "eu"), {Key: "token", Secret: true}}
	globals := NewVariablesImportItem("1", "globals", NewImportVariables(environment, nil)).CompletedVariables(ImportVariablesWrite{Globals: &workspace, KeptVariables: kept, Added: 2}, "L")
	if globals.Item.Status != "imported" || globals.Item.ID != "w1" || globals.Created.Variables != 2 || len(globals.Warnings) != 2 {
		t.Fatalf("globals outcome = %+v", globals)
	}
	if globals.Warnings[0].Message != "region keeps its value; the imported value was eu" || globals.Warnings[1].Message != "token keeps its value; the imported value was •••••" || globals.Warnings[0].Path != "Globals" || globals.Warnings[0].ItemKey != "1" {
		t.Errorf("warnings = %+v", globals.Warnings)
	}
	if !(ImportVariablesWrite{}).IsEmpty() || (ImportVariablesWrite{Globals: &workspace}).Scope() != NewGlobalsScopeRef() || (ImportVariablesWrite{Environment: &environment}).Scope() != environment.ScopeRef() {
		t.Error("scope and emptiness follow the planned write")
	}
}

func TestPostmanDumpSkippedParts(t *testing.T) {
	dump := PostmanDump{Skipped: []PostmanDumpPart{{Name: "mock", Files: 1}, {Name: "monitor", Files: 2}}}
	warnings := dump.SkippedWarnings()
	if len(warnings) != 2 || warnings[0].Path != "mock" || warnings[0].Message != "The dump's mock part (1 file) is not a collection or an environment and was skipped" || !strings.Contains(warnings[1].Message, "(2 files)") {
		t.Errorf("SkippedWarnings() = %+v", warnings)
	}
}

func TestWorkspaceKeepsImportedGlobalsPassthrough(t *testing.T) {
	workspace := NewWorkspace("w1", "Acme")
	workspace.Postman = SourceFields{"other": 1, "globals": "old"}
	updated := workspace.WithPostmanGlobals(SourceFields{"id": "g1"})
	if !reflect.DeepEqual(updated.Postman, SourceFields{"other": 1, "globals": map[string]any{"id": "g1"}}) || workspace.Postman["globals"] != "old" {
		t.Errorf("WithPostmanGlobals() = %+v", updated.Postman)
	}
	if unchanged := workspace.WithPostmanGlobals(nil); !reflect.DeepEqual(unchanged.Postman, workspace.Postman) {
		t.Error("no passthrough leaves the workspace unchanged")
	}
	result := NewImportResult().WithSourceWarnings([]ImportWarning{{Path: "mock"}})
	if len(result.Warnings) != 1 || len(NewImportSource("s", nil, nil, time.Time{}).WithSourceWarnings([]ImportWarning{{Path: "mock"}}).Warnings()) != 1 {
		t.Error("source warnings reach the source and the result")
	}
}
