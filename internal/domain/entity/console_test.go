package entity

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/consoleentrykind"
	"github.com/brienze1/xenon/internal/domain/enums/consolefilter"
	"github.com/brienze1/xenon/internal/domain/enums/consolelevel"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
	"github.com/brienze1/xenon/internal/domain/enums/teststatus"
)

var consoleTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func secretMask(text string) (string, bool) {
	masked := strings.ReplaceAll(text, "s3cret", MaskedValue)
	return masked, masked != text
}

func TestConsoleRequestEntry(t *testing.T) {
	request := EffectiveRequest{Method: "POST", URL: "http://x/a"}
	response := NewResponse(201, "Created", "HTTP/1.1", []KeyValue{}, []byte("{}"), 2, Timings{})
	script := preRef()
	entry := NewConsoleRequestEntry("01E", "01S", &script, true, request, response, consoleTime)
	if entry.Kind != consoleentrykind.Request || entry.Level != consolelevel.Info || entry.Summary != "POST http://x/a" || !entry.Nested || entry.Script != &script || entry.Request.URL != "http://x/a" || entry.Response.StatusCode != 201 || entry.SendID != "01S" || entry.ID != "01E" || !entry.CreatedAt.Equal(consoleTime) || entry.Arguments == nil {
		t.Fatalf("entry = %+v", entry)
	}
	if !entry.IsRequest() || entry.IsLog() || entry.IsError() {
		t.Errorf("a successful request is only a request")
	}
	for _, status := range []int{400, 404, 500} {
		failing := NewConsoleRequestEntry("1", "s", nil, false, request, NewResponse(status, "", "", nil, nil, 0, Timings{}), consoleTime)
		if failing.Level != consolelevel.Error || !failing.IsError() {
			t.Errorf("status %d is an error entry", status)
		}
	}
	if NewConsoleRequestEntry("1", "s", nil, false, request, NewResponse(399, "", "", nil, nil, 0, Timings{}), consoleTime).Level != consolelevel.Info {
		t.Errorf("status 399 is not an error")
	}
	failed := NewConsoleRequestEntry("1", "s", nil, false, request, NewFailedResponse(NewSendError(senderrorkind.Dns, "x"), Timings{}), consoleTime)
	if failed.Level != consolelevel.Error {
		t.Errorf("a failed send is an error entry")
	}
}

func TestConsoleTestFailureEntry(t *testing.T) {
	entry := NewConsoleTestFailureEntry("01E", "01S", TestResult{Name: "is 200", ErrorMessage: "expected 404", Script: postRef(), Status: teststatus.Failed}, consoleTime)
	want := ConsoleEntry{ID: "01E", SendID: "01S", Kind: consoleentrykind.TestFailure, Level: consolelevel.Error, Summary: "is 200: expected 404", Arguments: []string{"is 200", "expected 404"}, CreatedAt: consoleTime}
	if *entry.Script != postRef() {
		t.Fatalf("script = %+v", entry.Script)
	}
	entry.Script = nil
	if !reflect.DeepEqual(entry, want) {
		t.Errorf("entry = %+v", entry)
	}
	if !want.IsError() || want.IsLog() || want.IsRequest() {
		t.Errorf("a failed test is an error")
	}
}

func TestConsoleEntryKinds(t *testing.T) {
	cases := []struct {
		entry   ConsoleEntry
		log     bool
		isError bool
	}{
		{entry: ConsoleEntry{Kind: consoleentrykind.Log, Level: consolelevel.Log}, log: true},
		{entry: ConsoleEntry{Kind: consoleentrykind.Info, Level: consolelevel.Info}, log: true},
		{entry: ConsoleEntry{Kind: consoleentrykind.Log, Level: consolelevel.Error}, log: true, isError: true},
		{entry: ConsoleEntry{Kind: consoleentrykind.Error}, isError: true},
	}
	for _, tc := range cases {
		if tc.entry.IsLog() != tc.log || tc.entry.IsError() != tc.isError || tc.entry.IsRequest() {
			t.Errorf("%+v: log %v error %v", tc.entry, tc.entry.IsLog(), tc.entry.IsError())
		}
	}
}

func TestConsoleQueryFiltersAndSearches(t *testing.T) {
	script := postRef()
	request := EffectiveRequest{Method: "GET", URL: "http://x/Invoices"}
	entries := []ConsoleEntry{
		{ID: "log", Kind: consoleentrykind.Log, Level: consolelevel.Log, Summary: "Signing", Arguments: []string{"Signing"}, Script: &script},
		{ID: "warn", Kind: consoleentrykind.Log, Level: consolelevel.Warn, Summary: "short", Arguments: []string{"short", "detail TOKEN"}},
		{ID: "request", Kind: consoleentrykind.Request, Level: consolelevel.Info, Summary: "GET", Request: &request},
		{ID: "error", Kind: consoleentrykind.Error, Level: consolelevel.Error, Summary: "boom"},
	}
	cases := []struct {
		query ConsoleQuery
		want  string
	}{
		{query: ConsoleQuery{Limit: 10}, want: "log,warn,request,error"},
		{query: ConsoleQuery{Limit: 10, Filter: consolefilter.All}, want: "log,warn,request,error"},
		{query: ConsoleQuery{Limit: 10, Filter: consolefilter.Requests}, want: "request"},
		{query: ConsoleQuery{Limit: 10, Filter: consolefilter.Logs}, want: "log,warn"},
		{query: ConsoleQuery{Limit: 10, Filter: consolefilter.Errors}, want: "error"},
		{query: ConsoleQuery{Limit: 10, Search: "SIGN"}, want: "log"},
		{query: ConsoleQuery{Limit: 10, Search: "invoices"}, want: "request"},
		{query: ConsoleQuery{Limit: 10, Search: "token"}, want: "warn"},
		{query: ConsoleQuery{Limit: 10, Search: "login"}, want: ""},
		{query: ConsoleQuery{Limit: 10, Search: "upload"}, want: ""},
		{query: ConsoleQuery{Limit: 1, Filter: consolefilter.Logs, AfterID: "log"}, want: "warn"},
		{query: ConsoleQuery{Limit: 1, Filter: consolefilter.Errors}, want: "error"},
	}
	buffer := ConsoleBuffer{Entries: entries}
	for _, tc := range cases {
		if got := entryIDs(buffer.Find(tc.query)); got != tc.want {
			t.Errorf("%+v = %s, want %s", tc.query, got, tc.want)
		}
	}
	scriptName := ConsoleQuery{Limit: 10, Search: "upload"}
	script.OwnerName = "Upload"
	if got := entryIDs(buffer.Find(scriptName)); got != "log" {
		t.Errorf("search by script name = %s", got)
	}
}

func TestConsoleEntryMasksEverything(t *testing.T) {
	request := EffectiveRequest{
		Method:      "POST",
		URL:         "http://x/?k=s3cret",
		Headers:     []KeyValue{NewKeyValue("X-Key", "s3cret")},
		AutoHeaders: []AutoHeader{NewAutoHeader(HeaderAuthorization, "Bearer s3cret")},
		Body:        Body{Mode: bodymode.Raw, Raw: NewBodyRaw("text", "body s3cret")},
		Auth:        Auth{Bearer: AuthBearer{Token: "s3cret"}},
	}
	response := NewResponse(200, "OK", "HTTP/1.1", []KeyValue{NewKeyValue("X-Echo", "s3cret")}, []byte(`{"echo":"s3cret"}`), 17, Timings{})
	response.Error = &SendError{Message: "failed with s3cret"}
	entry := NewConsoleRequestEntry("1", "s", nil, false, request, response, consoleTime)
	entry.Values = []any{"s3cret", map[string]any{"s3cret": []any{"s3cret", 1.0}}, nil}
	masked := entry.WithSecretsMasked(secretMask)
	if !masked.Masked || masked.Summary != "POST http://x/?k="+MaskedValue || masked.Request.URL != "http://x/?k="+MaskedValue || masked.Request.Headers[0].Value != MaskedValue || masked.Request.AutoHeaders[0].Value != "Bearer "+MaskedValue || masked.Request.Body.Raw.Content != "body "+MaskedValue || masked.Request.Auth.Bearer.Token != MaskedValue {
		t.Errorf("request = %+v", masked.Request)
	}
	if masked.Response.Headers[0].Value != MaskedValue || string(masked.Response.Body) != `{"echo":"`+MaskedValue+`"}` || masked.Response.Error.Message != "failed with "+MaskedValue {
		t.Errorf("response = %+v", masked.Response)
	}
	wantValues := []any{MaskedValue, map[string]any{MaskedValue: []any{MaskedValue, 1.0}}, nil}
	if !reflect.DeepEqual(masked.Values, wantValues) {
		t.Errorf("values = %#v", masked.Values)
	}
	if entry.Request.URL != "http://x/?k=s3cret" || string(entry.Response.Body) != `{"echo":"s3cret"}` || entry.Response.Error.Message != "failed with s3cret" || entry.Values[0] != "s3cret" {
		t.Errorf("masking changed the original entry")
	}
	binary := NewResponse(200, "OK", "", []KeyValue{}, []byte("s3cret\x00"), 7, Timings{})
	binaryEntry := NewConsoleRequestEntry("1", "s", nil, false, EffectiveRequest{}, binary, consoleTime).WithSecretsMasked(secretMask)
	if string(binaryEntry.Response.Body) != "s3cret\x00" || binaryEntry.Masked {
		t.Errorf("binary bodies are not masked: %+v", binaryEntry)
	}
	if (ConsoleEntry{}).WithSecretsMasked(secretMask).Values != nil {
		t.Errorf("no values stay nil")
	}
}

func TestConsoleEntryRoutedBody(t *testing.T) {
	small := NewConsoleRequestEntry("01E", "s", nil, false, EffectiveRequest{}, NewResponse(200, "OK", "", []KeyValue{}, []byte("ok"), 2, Timings{}), consoleTime)
	if small.ServesBodyByRoute() || small.BodyKey() != "console:01E" || ConsoleBodyKey("x") != "console:x" {
		t.Errorf("small = %+v", small)
	}
	if string(small.WithoutRoutedBody().Response.Body) != "ok" {
		t.Errorf("a small body stays inline")
	}
	large := NewConsoleRequestEntry("01E", "s", nil, false, EffectiveRequest{}, NewResponse(200, "OK", "", []KeyValue{}, make([]byte, InlineBodyLimit+1), InlineBodyLimit+1, Timings{}), consoleTime)
	stripped := large.WithoutRoutedBody()
	if !large.ServesBodyByRoute() || stripped.Response.Body != nil || !stripped.ServesBodyByRoute() || large.Response.Body == nil {
		t.Errorf("large = %+v", stripped.Response)
	}
	failed := NewConsoleRequestEntry("1", "s", nil, false, EffectiveRequest{}, NewFailedResponse(NewSendError(senderrorkind.Dns, "x"), Timings{}), consoleTime)
	if failed.ServesBodyByRoute() || (ConsoleEntry{}).ServesBodyByRoute() {
		t.Errorf("no body to route")
	}
}

func TestTestSummary(t *testing.T) {
	summary := NewTestSummary([]TestResult{{Status: teststatus.Passed}, {Status: teststatus.Failed}, {Status: teststatus.Skipped}, {Status: teststatus.Passed}, {}})
	if summary != (TestSummary{Total: 5, Passed: 2, Failed: 1, Skipped: 1}) {
		t.Errorf("summary = %+v", summary)
	}
}

func TestBodyPreviewText(t *testing.T) {
	disabled := NewKeyValue("off", "1")
	disabled.Enabled = false
	file := NewFormDataRow("doc", "").WithType("file")
	file.Src = FileRef{Src: "/tmp/a.pdf"}
	disabledRow := NewFormDataRow("hidden", "x")
	disabledRow.Enabled = false
	cases := []struct {
		body Body
		want string
	}{
		{body: Body{Mode: bodymode.Raw, Raw: NewBodyRaw("json", `{"a":1}`)}, want: `{"a":1}`},
		{body: Body{Mode: bodymode.UrlEncoded, UrlEncoded: []KeyValue{NewKeyValue("a", "1"), disabled, NewKeyValue("b", "2")}}, want: "a=1&b=2"},
		{body: Body{Mode: bodymode.FormData, FormData: []FormDataRow{NewFormDataRow("name", "x"), file, disabledRow}}, want: "name: x\ndoc: @/tmp/a.pdf"},
		{body: Body{Mode: bodymode.GraphQl, GraphQl: BodyGraphQl{Query: "{ me }", Variables: `{"a":1}`}}, want: "{ me }\n\n{\"a\":1}"},
		{body: Body{Mode: bodymode.GraphQl, GraphQl: BodyGraphQl{Query: "{ me }"}}, want: "{ me }"},
		{body: Body{Mode: bodymode.Binary, Binary: FileRef{Src: "/tmp/b.bin"}}, want: "@/tmp/b.bin"},
		{body: DefaultBody(), want: ""},
	}
	for _, tc := range cases {
		if got := tc.body.PreviewText(); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.body.Mode, got, tc.want)
		}
	}
	if file.Type != formdatatype.File {
		t.Fatalf("file row = %+v", file)
	}
}

func TestEditedScripts(t *testing.T) {
	stored := NewRequest("01R", "Create")
	stored.Scripts = Scripts{PreRequest: "saved"}
	stored.InheritScripts = false
	draft := NewRequest("01R", "Create")
	if got := draft.WithScriptsOf(stored); got.Scripts.PreRequest != "saved" || got.InheritScripts {
		t.Errorf("a draft without scripts takes the stored ones: %+v", got.Scripts)
	}
	draft.Scripts = Scripts{PreRequest: "edited"}
	draft.ScriptsEdited = true
	if got := draft.WithScriptsOf(stored); got.Scripts.PreRequest != "edited" || !got.InheritScripts {
		t.Errorf("edited scripts win: %+v", got.Scripts)
	}
	folder := Collection{Scripts: Scripts{PostResponse: "saved"}, InheritScripts: true}
	kept := folder.WithSettings(Collection{Description: "d"})
	if kept.Scripts.PostResponse != "saved" || !kept.InheritScripts || kept.ScriptsEdited || kept.Description != "d" {
		t.Errorf("settings without scripts keep them: %+v", kept)
	}
	edited := folder.WithSettings(Collection{Scripts: Scripts{PreRequest: "new"}, ScriptsEdited: true})
	if edited.Scripts.PreRequest != "new" || edited.Scripts.PostResponse != "" || edited.InheritScripts || !edited.ScriptsEdited {
		t.Errorf("edited settings replace the scripts: %+v", edited)
	}
}
