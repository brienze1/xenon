package dto

import (
	"reflect"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

func writtenRequest() entity.Request {
	request := entity.NewRequest("01R", "Create")
	request.Method = "GET"
	request.URL = "{{api}}/items?a=1"
	request.Headers = []entity.KeyValue{entity.NewKeyValue("X-On", "1"), {Key: "X-Off", Value: "0"}}
	return request
}

func TestNewScriptRequestShowsThePostmanShape(t *testing.T) {
	request := writtenRequest()
	request.Body = entity.Body{Mode: bodymode.Raw, Raw: entity.BodyRaw{Language: rawlanguage.Xml, Content: "<a/>"}}
	want := ScriptRequest{
		Method: "GET", URL: "{{api}}/items?a=1",
		Header: []ScriptKeyValue{{Key: "X-On", Value: "1"}, {Key: "X-Off", Value: "0", Disabled: true}},
		Body:   &ScriptBody{Mode: "raw", Raw: "<a/>"},
	}
	if got := NewScriptRequest(request); !reflect.DeepEqual(got, want) {
		t.Fatalf("NewScriptRequest = %+v", got)
	}
	form := entity.DefaultBody()
	form.Mode = bodymode.FormData
	form.FormData = []entity.FormDataRow{entity.NewFormDataRow("f", "v"), {Key: "off", Value: "x"}}
	if got := NewScriptBody(form); !reflect.DeepEqual(got, ScriptBody{Mode: "formdata", FormData: []ScriptKeyValue{{Key: "f", Value: "v"}, {Key: "off", Value: "x", Disabled: true}}}) {
		t.Errorf("formdata = %+v", got)
	}
	encoded := entity.DefaultBody()
	encoded.Mode = bodymode.UrlEncoded
	encoded.UrlEncoded = []entity.KeyValue{entity.NewKeyValue("k", "v")}
	if got := NewScriptBody(encoded); !reflect.DeepEqual(got, ScriptBody{Mode: "urlencoded", UrlEncoded: []ScriptKeyValue{{Key: "k", Value: "v"}}}) {
		t.Errorf("urlencoded = %+v", got)
	}
	if got := NewScriptBody(entity.Body{Mode: bodymode.GraphQl, GraphQl: entity.BodyGraphQl{Query: "{ a }", Variables: `{"v":1}`}}); !reflect.DeepEqual(got, ScriptBody{Mode: "graphql", GraphQl: &ScriptGraphQl{Query: "{ a }", Variables: `{"v":1}`}}) {
		t.Errorf("graphql = %+v", got)
	}
	if got := NewScriptRequest(entity.NewRequest("", "")); got.Body != nil || got.Auth != nil {
		t.Errorf("no body and an inherited auth are left out: %+v", got)
	}
}

func TestScriptRequestAppliesChangesToTheWrittenRequest(t *testing.T) {
	written := writtenRequest()
	written.Body = entity.Body{Mode: bodymode.GraphQl, GraphQl: entity.BodyGraphQl{Query: "{ a }"}}
	view := NewScriptRequest(written)
	view.Method = "POST"
	view.URL = "{{api}}/items?a=1&b=2"
	view.Header = append(view.Header, ScriptKeyValue{Key: "X-Count", Value: 3.0})
	changed := view.ToEntity(written)
	if changed.Method != "POST" || changed.URL != "{{api}}/items?a=1&b=2" || changed.ID != "01R" || changed.Body.GraphQl.Query != "{ a }" {
		t.Fatalf("changed = %+v", changed)
	}
	wantHeaders := []entity.KeyValue{entity.NewKeyValue("X-On", "1"), {Key: "X-Off", Value: "0"}, entity.NewKeyValue("X-Count", "3")}
	if !reflect.DeepEqual(changed.Headers, wantHeaders) {
		t.Errorf("headers = %+v", changed.Headers)
	}
	if written.Method != "GET" || len(written.Headers) != 2 {
		t.Errorf("the written request changed: %+v", written)
	}
}

func TestScriptRequestBodyChanges(t *testing.T) {
	jsonBody := writtenRequest()
	jsonBody.Body = entity.Body{Mode: bodymode.Raw, Raw: entity.BodyRaw{Language: rawlanguage.Json, Content: "{}"}}
	view := NewScriptRequest(jsonBody)
	view.Body.Raw = `{"a":1}`
	if got := view.ToEntity(jsonBody).Body; got.Mode != bodymode.Raw || got.Raw != (entity.BodyRaw{Language: rawlanguage.Json, Content: `{"a":1}`}) {
		t.Errorf("raw keeps its language: %+v", got)
	}
	empty := writtenRequest()
	view = NewScriptRequest(empty)
	view.Body = &ScriptBody{Mode: "raw", Raw: "text"}
	if got := view.ToEntity(empty).Body; got.Raw != (entity.BodyRaw{Language: rawlanguage.Text, Content: "text"}) || got.Mode != bodymode.Raw {
		t.Errorf("a new raw body is text: %+v", got)
	}
	view.Body = &ScriptBody{Mode: "urlencoded", UrlEncoded: []ScriptKeyValue{{Key: "k", Value: "v"}, {Key: "n", Value: nil, Disabled: true}}}
	if got := view.ToEntity(empty).Body; got.Mode != bodymode.UrlEncoded || !reflect.DeepEqual(got.UrlEncoded, []entity.KeyValue{entity.NewKeyValue("k", "v"), {Key: "n"}}) {
		t.Errorf("urlencoded = %+v", got)
	}
	view.Body = &ScriptBody{Mode: "formdata", FormData: []ScriptKeyValue{{Key: "f", Value: "v", Disabled: true}}}
	if got := view.ToEntity(empty).Body; got.Mode != bodymode.FormData || len(got.FormData) != 1 || got.FormData[0].Type != formdatatype.Text || got.FormData[0].Enabled || got.FormData[0].Value != "v" {
		t.Errorf("formdata = %+v", got)
	}
	view.Body = &ScriptBody{Mode: "file"}
	if got := view.ToEntity(empty).Body; got.Mode != bodymode.None {
		t.Errorf("an unknown mode has no body: %+v", got)
	}
	files := writtenRequest()
	files.Body = entity.Body{Mode: bodymode.FormData, FormData: []entity.FormDataRow{{Key: "f", Type: formdatatype.File, Src: entity.FileRef{Src: "a.bin"}, Enabled: true}}}
	if got := NewScriptRequest(files).ToEntity(files).Body; !reflect.DeepEqual(got, files.Body) {
		t.Errorf("an unchanged body keeps its files: %+v", got)
	}
}

func TestScriptBodyEquals(t *testing.T) {
	base := ScriptBody{Mode: "urlencoded", UrlEncoded: []ScriptKeyValue{{Key: "k", Value: "1"}}, FormData: []ScriptKeyValue{{Key: "f", Value: "v"}}}
	cases := map[string]ScriptBody{
		"mode":           {Mode: "raw", UrlEncoded: base.UrlEncoded, FormData: base.FormData},
		"raw":            {Mode: base.Mode, Raw: "x", UrlEncoded: base.UrlEncoded, FormData: base.FormData},
		"length":         {Mode: base.Mode, UrlEncoded: nil, FormData: base.FormData},
		"key":            {Mode: base.Mode, UrlEncoded: []ScriptKeyValue{{Key: "other", Value: "1"}}, FormData: base.FormData},
		"value":          {Mode: base.Mode, UrlEncoded: []ScriptKeyValue{{Key: "k", Value: "2"}}, FormData: base.FormData},
		"disabled":       {Mode: base.Mode, UrlEncoded: []ScriptKeyValue{{Key: "k", Value: "1", Disabled: true}}, FormData: base.FormData},
		"form data rows": {Mode: base.Mode, UrlEncoded: base.UrlEncoded, FormData: []ScriptKeyValue{{Key: "f", Value: "w"}}},
	}
	for name, other := range cases {
		if base.Equals(other) {
			t.Errorf("%s differs but Equals is true", name)
		}
	}
	if !base.Equals(ScriptBody{Mode: "urlencoded", UrlEncoded: []ScriptKeyValue{{Key: "k", Value: 1.0}}, FormData: []ScriptKeyValue{{Key: "f", Value: "v"}}}) {
		t.Errorf("values compare as text")
	}
}

func TestNewScriptResponse(t *testing.T) {
	response := entity.NewResponse(201, "Created", "HTTP/1.1", []entity.KeyValue{entity.NewKeyValue("Content-Type", "application/json")}, []byte(`{"a":1}`), 7, entity.Timings{Total: 12 * time.Millisecond})
	want := ScriptResponse{Code: 201, Status: "Created", Header: []ScriptKeyValue{{Key: "Content-Type", Value: "application/json"}}, Body: `{"a":1}`, ResponseTime: 12, ResponseSize: 7}
	if got := NewScriptResponse(response); !reflect.DeepEqual(got, want) {
		t.Errorf("NewScriptResponse = %+v", got)
	}
}

func TestScriptRequestAuthAndGraphQl(t *testing.T) {
	written := writtenRequest()
	written.Auth = entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: "t1"}}
	view := NewScriptRequest(written)
	if view.Auth == nil || view.Auth.Type != "bearer" || view.Auth.Bearer[0].Value != "t1" {
		t.Fatalf("auth = %+v", view.Auth)
	}
	if got := view.ToEntity(written).Auth; got.Bearer.Token != "t1" {
		t.Errorf("unchanged auth = %+v", got)
	}
	view.Auth = &ScriptAuth{Type: "basic", Basic: []ScriptKeyValue{{Key: "username", Value: "u"}, {Key: "password", Value: "p"}}}
	if got := view.ToEntity(written).Auth; got.Type != authtype.Basic || got.Basic != (entity.AuthBasic{Username: "u", Password: "p"}) {
		t.Errorf("basic = %+v", got)
	}
	view.Auth = &ScriptAuth{Type: "none"}
	if got := view.ToEntity(written).Auth; got.Type != authtype.None {
		t.Errorf("none = %+v", got)
	}
	view.Auth = &ScriptAuth{Type: "oauth2"}
	if got := view.ToEntity(written).Auth; got.Type != authtype.Bearer {
		t.Errorf("an auth type Xenon cannot send keeps the written auth: %+v", got)
	}
	if (&ScriptAuth{Type: "none"}).Equals(nil) {
		t.Errorf("nil auth differs")
	}
	none := writtenRequest()
	none.Auth = entity.Auth{Type: authtype.None}
	if got := NewScriptRequest(none).Auth; got == nil || got.Type != "none" {
		t.Errorf("none = %+v", got)
	}
	basic := writtenRequest()
	basic.Auth = entity.Auth{Type: authtype.Basic, Basic: entity.AuthBasic{Username: "a"}}
	if got := NewScriptRequest(basic).Auth; got == nil || len(got.Basic) != 2 {
		t.Errorf("basic = %+v", got)
	}
	graph := writtenRequest()
	graph.Body = entity.Body{Mode: bodymode.GraphQl, GraphQl: entity.BodyGraphQl{Query: "{ a }"}}
	view = NewScriptRequest(graph)
	view.Body.GraphQl = &ScriptGraphQl{Query: "{ b }", Variables: map[string]any{"x": 1.0}}
	if got := view.ToEntity(graph).Body; got.Mode != bodymode.GraphQl || got.GraphQl != (entity.BodyGraphQl{Query: "{ b }", Variables: `{"x":1}`}) {
		t.Errorf("graphql = %+v", got)
	}
	if !(*ScriptGraphQl)(nil).Equals(nil) || (&ScriptGraphQl{}).Equals(nil) {
		t.Errorf("graphql equality with nil")
	}
}
