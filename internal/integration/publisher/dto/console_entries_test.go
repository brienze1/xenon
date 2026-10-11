package dto

import (
	"reflect"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/scriptownerkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
)

func TestConsoleEntriesEvent(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	script := entity.ScriptRef{Phase: scriptphase.PreRequest, OwnerKind: scriptownerkind.Collection, ItemID: "01C", OwnerName: "Billing"}
	response := entity.NewResponse(200, "OK", "", nil, []byte("ok"), 2, entity.Timings{Total: 2500 * time.Microsecond})
	request := entity.NewConsoleRequestEntry("01E", "01S", &script, true, entity.EffectiveRequest{Method: "GET", URL: "http://x"}, response, at)
	failed := entity.NewConsoleRequestEntry("01F", "01S", nil, false, entity.EffectiveRequest{Method: "GET", URL: "http://y"}, entity.NewFailedResponse(entity.NewSendError(senderrorkind.Dns, "y"), entity.Timings{}), at)
	scriptError := entity.NewConsoleErrorEntry("01G", "01S", entity.ScriptError{Script: script, Message: "Error: x", Line: 2, Column: 4}, at)
	event := NewConsoleEntries([]entity.ConsoleEntry{request, failed, scriptError})
	want := ConsoleEntry{
		Script:    &ConsoleScript{OwnerKind: "collection", OwnerId: "01C", OwnerName: "Billing", Phase: "preRequest", Label: "Billing · Pre-request"},
		Request:   &ConsoleRequest{Method: "GET", Url: "http://x", Nested: true},
		Response:  &ConsoleResponse{StatusCode: 200, StatusText: "OK", SizeBytes: 2, TimeMs: 2.5},
		Id:        "01E",
		SendId:    "01S",
		Kind:      "request",
		Level:     "info",
		Summary:   "GET http://x",
		CreatedAt: "2026-01-01T00:00:00.000Z",
		Arguments: []string{},
	}
	if event.Cleared || len(event.Entries) != 3 || !reflect.DeepEqual(event.Entries[0], want) {
		t.Fatalf("event = %+v", event.Entries[0])
	}
	if got := event.Entries[1]; got.Script != nil || got.Response.Error == "" || got.Level != "error" || got.Line != nil {
		t.Errorf("failed = %+v", got)
	}
	if got := event.Entries[2]; *got.Line != 2 || *got.Column != 4 || got.Request != nil || got.Response != nil {
		t.Errorf("error = %+v", got)
	}
	cleared := NewConsoleCleared()
	if !cleared.Cleared || cleared.Entries == nil || len(cleared.Entries) != 0 {
		t.Errorf("cleared = %+v", cleared)
	}
}
