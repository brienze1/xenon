package dto

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
)

func decodeRequest(t *testing.T, raw string) Request {
	t.Helper()
	var request Request
	if err := json.Unmarshal([]byte(raw), &request); err != nil {
		t.Fatal(err)
	}
	return request
}

func TestEmptyRequestGetsTheDefaults(t *testing.T) {
	request := decodeRequest(t, `{}`)
	got := request.ToEntity()
	want := entity.NewRequest("", "")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ToEntity = %+v, want %+v", got, want)
	}
}

func TestRequestToEntity(t *testing.T) {
	request := decodeRequest(t, `{
		"id": "01J", "name": " Create ", "description": "d", "method": "purge", "url": "https://h/:id?a=1",
		"params": [ { "key": "a", "value": "1", "description": "x" }, { "key": "b", "value": "2", "enabled": false } ],
		"path_variables": [ { "key": "id", "value": "7", "description": "v" } ],
		"headers": [ { "key": "H", "value": "1", "enabled": true } ],
		"auto_headers": { "disabled": ["Accept"] },
		"auth": { "type": "basic", "basic": { "username": "u", "password": "p" }, "bearer": { "token": "t" } },
		"body": {
			"mode": "formdata",
			"raw": { "language": "xml", "content": "<a/>" },
			"formdata": [ { "key": "f", "type": "file", "src": "a.bin" }, { "key": "t", "value": "v", "enabled": false } ],
			"urlencoded": [ { "key": "u", "value": "1" } ],
			"binary": { "src": "b.bin" },
			"graphql": { "query": "q", "variables": "{}" }
		},
		"settings": { "timeout_ms": 10, "follow_redirects": false, "verify_ssl": false, "encode_url": false }
	}`)
	got := request.ToEntity()
	want := entity.Request{
		ID: "01J", Name: "Create", Description: "d", Method: "PURGE", URL: "https://h/:id?a=1",
		Params:              []entity.KeyValue{{Key: "a", Value: "1", Enabled: true, Description: "x"}, {Key: "b", Value: "2"}},
		PathVariables:       []entity.PathVariable{{Key: "id", Value: "7", Description: "v"}},
		Headers:             []entity.KeyValue{entity.NewKeyValue("H", "1")},
		DisabledAutoHeaders: []string{"Accept"},
		Auth:                entity.Auth{Type: authtype.Basic, Basic: entity.AuthBasic{Username: "u", Password: "p"}, Bearer: entity.AuthBearer{Token: "t"}},
		Body: entity.Body{
			Mode: bodymode.FormData,
			Raw:  entity.BodyRaw{Language: rawlanguage.Xml, Content: "<a/>"},
			FormData: []entity.FormDataRow{
				{Key: "f", Type: formdatatype.File, Src: entity.FileRef{Src: "a.bin"}, Enabled: true},
				{Key: "t", Type: formdatatype.Text, Value: "v"},
			},
			UrlEncoded: []entity.KeyValue{entity.NewKeyValue("u", "1")},
			Binary:     entity.FileRef{Src: "b.bin"},
			GraphQl:    entity.BodyGraphQl{Query: "q", Variables: "{}"},
		},
		Settings: entity.RequestSettings{TimeoutMs: 10},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ToEntity =\n%+v\nwant\n%+v", got, want)
	}
	back := NewRequest(got)
	if !reflect.DeepEqual(back.ToEntity(), got) {
		t.Fatalf("NewRequest must round trip: %+v", back)
	}
	if back.Method != "PURGE" || back.Settings.TimeoutMs != 10 || *back.Settings.VerifySsl || *back.Params[1].Enabled {
		t.Fatalf("NewRequest = %+v", back)
	}
	if !reflect.DeepEqual(back.LogValue(), map[string]any{"id": "01J", "method": "PURGE", "url": "https://h/:id?a=1"}) {
		t.Fatalf("LogValue = %v", back.LogValue())
	}
}

func TestCommandLogValues(t *testing.T) {
	request := Request{Id: "1", Method: "GET", Url: "u"}
	cases := []struct {
		got  any
		want any
	}{
		{ApplyRequest{Request: request}.LogValue(), request.LogValue()},
		{AutoHeadersRequest{LocalId: "ws", Request: request}.LogValue(), map[string]any{"local_id": "ws", "request": request.LogValue()}},
		{SendRequest{LocalId: "ws", TabId: "t", Request: request}.LogValue(), map[string]any{"local_id": "ws", "tab_id": "t", "request": request.LogValue()}},
		{NewBulkText("abc").LogValue(), map[string]any{"length": 3}},
		{BulkRows{Rows: []KeyValue{{}, {}}}.LogValue(), map[string]any{"rows": 2}},
		{NewAutoHeaderList([]entity.AutoHeader{{Key: "A"}}).LogValue(), map[string]any{"headers": 1}},
		{SendResult{TabId: "t", Ok: true, StatusCode: 200, SizeBytes: 5}.LogValue(), map[string]any{"tab_id": "t", "ok": true, "status_code": 200, "size_bytes": int64(5)}},
	}
	for _, tc := range cases {
		if !reflect.DeepEqual(tc.got, tc.want) {
			t.Errorf("LogValue = %v, want %v", tc.got, tc.want)
		}
	}
}

func TestBulkAndAutoHeaders(t *testing.T) {
	rows := []entity.KeyValue{entity.NewKeyValue("a", "1"), {Key: "b"}}
	bulk := NewBulkRows(rows)
	if !reflect.DeepEqual(bulk.ToEntity(), rows) {
		t.Fatalf("bulk rows = %+v", bulk.ToEntity())
	}
	list := NewAutoHeaderList([]entity.AutoHeader{{Key: "A", Value: "1", Disabled: true, Computed: true}})
	if !reflect.DeepEqual(list, AutoHeaderList{{Key: "A", Value: "1", Disabled: true}}) {
		t.Fatalf("auto headers = %+v", list)
	}
	if len(NewAutoHeaderList(nil)) != 0 || NewAutoHeaderList(nil) == nil {
		t.Fatal("no auto headers is an empty list")
	}
}

func TestNewSendResult(t *testing.T) {
	timings := entity.Timings{Dns: 1500 * time.Microsecond, Connect: 2 * time.Millisecond, Tls: 3 * time.Millisecond, FirstByte: 4 * time.Millisecond, Download: 5 * time.Millisecond, Total: 15 * time.Millisecond}
	headers := []entity.KeyValue{entity.NewKeyValue("Content-Type", "text/plain")}
	ok := NewSendResult("tab", entity.NewResponse(200, "OK", "HTTP/1.1", headers, []byte("hi"), 2, timings))
	if !ok.Ok || ok.Error != nil || *ok.InlineBody != "hi" || ok.BodyUrl != "/xenon/response/tab" || ok.ContentType != "text/plain" || ok.IsBinary {
		t.Fatalf("ok = %+v", ok)
	}
	if !reflect.DeepEqual(ok.Headers, []ResponseHeader{{Key: "Content-Type", Value: "text/plain"}}) || ok.StatusCode != 200 || ok.StatusText != "OK" || ok.Proto != "HTTP/1.1" || ok.SizeBytes != 2 || ok.TabId != "tab" {
		t.Fatalf("ok = %+v", ok)
	}
	if ok.Timings != (SendResultTimings{DnsMs: 1.5, ConnectMs: 2, TlsMs: 3, FirstByteMs: 4, DownloadMs: 5, TotalMs: 15}) {
		t.Fatalf("timings = %+v", ok.Timings)
	}
	binary := NewSendResult("tab", entity.NewResponse(200, "OK", "HTTP/1.1", nil, []byte{0}, 1, timings))
	if binary.InlineBody != nil || !binary.IsBinary {
		t.Fatalf("binary = %+v", binary)
	}
	failed := NewSendResult("tab", entity.NewFailedResponse(entity.NewSendError(senderrorkind.Timeout, "5"), timings))
	if failed.Ok || failed.InlineBody != nil || *failed.Error != (SendResultError{Kind: "timeout", Message: "Timed out after 5 ms", Hint: "Increase the timeout in the request Settings (0 = no timeout)"}) || failed.Headers == nil {
		t.Fatalf("failed = %+v", failed)
	}
}

func TestResponseBodyAndSavePath(t *testing.T) {
	response := entity.NewResponse(200, "OK", "HTTP/1.1", []entity.KeyValue{entity.NewKeyValue("Content-Type", "image/png")}, nil, 0, entity.Timings{})
	if body := NewResponseBody(response, nil); body.ContentType != "image/png" {
		t.Fatalf("body = %+v", body)
	}
	save := SaveResponseBody{Path: "/tmp/a/../b.bin"}
	if save.GetPath() != filepath.FromSlash("/tmp/b.bin") {
		t.Fatalf("path = %q", save.GetPath())
	}
}
