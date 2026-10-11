package dto

import (
	"reflect"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/consolefilter"
	"github.com/brienze1/xenon/internal/domain/enums/consolelevel"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptownerkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
)

func TestListConsoleToEntity(t *testing.T) {
	list := ListConsole{AfterId: "01A", Filter: "errors", Search: "x", Limit: 5}
	if got := list.ToEntity(); got != (entity.ConsoleQuery{AfterID: "01A", Filter: consolefilter.Errors, Search: "x", Limit: 5}) {
		t.Errorf("query = %+v", got)
	}
	empty := ListConsole{Limit: 1}
	if got := empty.ToEntity(); got.Filter != consolefilter.All {
		t.Errorf("filter = %q", got.Filter)
	}
}

func TestConsoleRequestEntryDto(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	script := entity.ScriptRef{Phase: scriptphase.PreRequest, OwnerKind: scriptownerkind.Request, ItemID: "01R", OwnerName: "Create"}
	request := entity.EffectiveRequest{Method: "POST", URL: "http://x/a", Headers: []entity.KeyValue{entity.NewKeyValue("X-A", "1")}, AutoHeaders: []entity.AutoHeader{entity.NewAutoHeader("Accept", "*/*")}, Body: entity.Body{Mode: bodymode.Raw, Raw: entity.NewBodyRaw("text", "hi")}}
	response := entity.NewResponse(201, "Created", "HTTP/1.1", []entity.KeyValue{entity.NewKeyValue("Content-Type", "text/plain")}, []byte("done"), 4, entity.Timings{Total: 1500 * time.Microsecond})
	entry := entity.NewConsoleRequestEntry("01E", "01S", &script, true, request, response, at)
	summary := NewConsoleEntry(entry)
	wantSummary := ConsoleEntry{
		Script:    &ScriptRef{OwnerKind: "request", OwnerId: "01R", OwnerName: "Create", Phase: "preRequest", Label: "Create · Pre-request"},
		Request:   &ConsoleRequest{Method: "POST", Url: "http://x/a", Nested: true},
		Response:  &ConsoleResponse{StatusCode: 201, StatusText: "Created", SizeBytes: 4, TimeMs: 1.5},
		Id:        "01E",
		SendId:    "01S",
		Kind:      "request",
		Level:     "info",
		Summary:   "POST http://x/a",
		CreatedAt: "2026-01-01T00:00:00.000Z",
		Arguments: []string{},
	}
	if !reflect.DeepEqual(summary, wantSummary) {
		t.Fatalf("summary = %+v", summary)
	}
	detail := NewConsoleEntryDetail(entry)
	body := "hi"
	wantRequest := ConsoleRequest{Method: "POST", Url: "http://x/a", Nested: true, Headers: []ResponseHeader{{Key: "X-A", Value: "1"}, {Key: "Accept", Value: "*/*"}}, Body: &body}
	if !reflect.DeepEqual(*detail.Request, wantRequest) {
		t.Errorf("request = %+v", *detail.Request)
	}
	responseBody := "done"
	wantResponse := ConsoleResponse{StatusCode: 201, StatusText: "Created", SizeBytes: 4, TimeMs: 1.5, ContentType: "text/plain", Headers: []ResponseHeader{{Key: "Content-Type", Value: "text/plain"}}, Body: &responseBody}
	if !reflect.DeepEqual(*detail.Response, wantResponse) {
		t.Errorf("response = %+v", *detail.Response)
	}
	if detail.Values != nil || detail.Line != nil {
		t.Errorf("detail = %+v", detail)
	}
}

func TestConsoleEntryDetailBodies(t *testing.T) {
	large := entity.NewResponse(200, "OK", "", []entity.KeyValue{}, make([]byte, entity.InlineBodyLimit+1), entity.InlineBodyLimit+1, entity.Timings{})
	detail := NewConsoleEntryDetail(entity.NewConsoleRequestEntry("01E", "s", nil, false, entity.EffectiveRequest{Method: "GET"}, large, time.Time{}).WithoutRoutedBody())
	if detail.Response.Body != nil || detail.Response.BodyUrl != "/xenon/response/console:01E" || detail.Request.Body != nil || detail.Script != nil {
		t.Errorf("large = %+v / %+v", detail.Response, detail.Request)
	}
	failed := entity.NewFailedResponse(entity.NewSendError(senderrorkind.Dns, "x.invalid"), entity.Timings{})
	failedDetail := NewConsoleEntryDetail(entity.NewConsoleRequestEntry("01E", "s", nil, false, entity.EffectiveRequest{}, failed, time.Time{}))
	if failedDetail.Response.Body != nil || failedDetail.Response.BodyUrl != "" || failedDetail.Response.Error == nil || failedDetail.Response.Error.Kind != "dns" {
		t.Errorf("failed = %+v", failedDetail.Response)
	}
}

func TestConsoleLogAndErrorEntryDto(t *testing.T) {
	script := entity.ScriptRef{Phase: scriptphase.PostResponse, OwnerKind: scriptownerkind.Folder, OwnerName: "Auth"}
	log := entity.NewConsoleLogEntry("01E", "s", script, entity.NewScriptLog(consolelevel.Info, []string{"a"}, []any{map[string]any{"k": 1.0}}), time.Time{})
	if summary := NewConsoleEntry(log); summary.Values != nil || summary.Request != nil || summary.Response != nil || summary.Level != "info" {
		t.Errorf("summary = %+v", summary)
	}
	if detail := NewConsoleEntryDetail(log); !reflect.DeepEqual(detail.Values, []any{map[string]any{"k": 1.0}}) {
		t.Errorf("values = %+v", detail.Values)
	}
	scriptError := entity.NewConsoleErrorEntry("01E", "s", entity.ScriptError{Script: script, Message: "Error: x", Line: 3, Column: 9}, time.Time{})
	if got := NewConsoleEntry(scriptError); *got.Line != 3 || *got.Column != 9 || got.Kind != "error" {
		t.Errorf("error = %+v", got)
	}
	if got := NewConsoleEntries([]entity.ConsoleEntry{log, scriptError}); len(got) != 2 || got[1].Id != "01E" {
		t.Errorf("entries = %+v", got)
	}
	if got := NewConsoleEntries(nil); got == nil || len(got) != 0 {
		t.Errorf("no entries = %+v", got)
	}
}

func TestRequestScripts(t *testing.T) {
	inherit := false
	edited := Request{Name: "x", Scripts: &Scripts{PreRequest: "a", PostResponse: "b", Disabled: []string{"postResponse"}}, InheritScripts: &inherit}
	request := edited.ToEntity()
	if !request.ScriptsEdited || request.Scripts.PreRequest != "a" || request.Scripts.PostResponse != "b" || len(request.Scripts.Disabled) != 1 || request.Scripts.Disabled[0] != scriptphase.PostResponse || request.InheritScripts {
		t.Fatalf("request = %+v", request)
	}
	defaulted := Request{Scripts: &Scripts{}}
	if got := defaulted.ToEntity(); !got.InheritScripts || !got.ScriptsEdited || got.Scripts.Disabled != nil {
		t.Errorf("inherit defaults to true: %+v", got)
	}
	plain := Request{}
	if got := plain.ToEntity(); got.ScriptsEdited || !got.InheritScripts {
		t.Errorf("no scripts = %+v", got)
	}
	back := NewRequest(request)
	if back.Scripts.PreRequest != "a" || back.Scripts.Disabled[0] != "postResponse" || *back.InheritScripts {
		t.Errorf("dto = %+v", back.Scripts)
	}
	if empty := NewScripts(entity.Scripts{}); empty.Disabled == nil {
		t.Errorf("disabled is never nil")
	}
	settings := SaveItemSettings{Id: "01F", Scripts: &Scripts{PreRequest: "p"}}
	if got := settings.ToEntity(); !got.ScriptsEdited || got.Scripts.PreRequest != "p" || !got.InheritScripts {
		t.Errorf("settings = %+v", got)
	}
	plainSettings := SaveItemSettings{Id: "01F"}
	if got := plainSettings.ToEntity(); got.ScriptsEdited {
		t.Errorf("settings without scripts = %+v", got)
	}
	folder := NewItemSettings(entity.Collection{Kind: itemkind.Folder, InheritScripts: true})
	collection := NewItemSettings(entity.Collection{Kind: itemkind.Collection, InheritScripts: true})
	if folder.InheritScripts == nil || !*folder.InheritScripts || collection.InheritScripts != nil || collection.Scripts == nil {
		t.Errorf("folder %+v collection %+v", folder, collection)
	}
}
