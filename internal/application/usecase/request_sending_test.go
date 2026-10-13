package usecase

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

func bearer(token string) entity.Auth {
	return entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: token}}
}

func folder(auth entity.Auth, headers ...entity.KeyValue) entity.Collection {
	container := entity.NewFolder("f", "F")
	container.Auth = auth
	container.Headers = headers
	return container
}

func TestResolveEffectiveRequestUrl(t *testing.T) {
	cases := []struct {
		url    string
		encode bool
		want   string
	}{
		{"api.test/users/:id?q=a b#frag", true, "http://api.test/users/42?q=a%20b"},
		{"api.test/users/:id?q=a b#frag", false, "http://api.test/users/42?q=a b"},
		{"", true, ""},
	}
	for _, tc := range cases {
		request := entity.NewRequest("", "r")
		request.Method = " purge "
		request.URL = tc.url
		request.PathVariables = []entity.PathVariable{{Key: "id", Value: "42"}}
		request.Settings.EncodeUrl = tc.encode
		effective := ResolveEffectiveRequest(request, nil, "")
		if effective.URL != tc.want || effective.Method != "PURGE" {
			t.Errorf("%q encode=%v: %s %q, want %q", tc.url, tc.encode, effective.Method, effective.URL, tc.want)
		}
	}
}

func TestResolveEffectiveRequestAuth(t *testing.T) {
	collection := entity.NewCollection("c", "C")
	collection.Auth = bearer("collection")
	cases := map[string]struct {
		request   entity.Auth
		ancestors []entity.Collection
		want      entity.Auth
	}{
		"own auth wins":           {bearer("own"), []entity.Collection{collection}, bearer("own")},
		"nearest ancestor wins":   {entity.DefaultAuth(), []entity.Collection{collection, folder(bearer("folder")), folder(entity.DefaultAuth())}, bearer("folder")},
		"collection is the root":  {entity.DefaultAuth(), []entity.Collection{collection, folder(entity.DefaultAuth())}, bearer("collection")},
		"inherit at root is none": {entity.DefaultAuth(), []entity.Collection{folder(entity.DefaultAuth())}, entity.Auth{Type: authtype.None}},
		"no ancestors is none":    {entity.DefaultAuth(), nil, entity.Auth{Type: authtype.None}},
		"explicit none stays":     {entity.Auth{Type: authtype.None}, []entity.Collection{collection}, entity.Auth{Type: authtype.None}},
	}
	for name, tc := range cases {
		request := entity.NewRequest("", "r")
		request.Auth = tc.request
		if got := ResolveEffectiveRequest(request, tc.ancestors, "").Auth; !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: auth = %+v, want %+v", name, got, tc.want)
		}
	}
}

func TestResolveEffectiveRequestHeaders(t *testing.T) {
	collection := entity.NewCollection("c", "C")
	collection.Headers = []entity.KeyValue{entity.NewKeyValue("X-Env", "prod"), entity.NewKeyValue("X-Team", "core"), {Key: "X-Off", Value: "1"}}
	ancestors := []entity.Collection{collection, folder(entity.DefaultAuth(), entity.NewKeyValue("x-env", "staging"), entity.NewKeyValue("X-Folder", "f"))}
	request := entity.NewRequest("", "r")
	request.Headers = []entity.KeyValue{entity.NewKeyValue("X-ENV", "dev"), {Key: "X-Team", Value: "ignored"}, entity.NewKeyValue(" ", "blank")}
	got := ResolveEffectiveRequest(request, ancestors, "").Headers
	want := []entity.KeyValue{entity.NewKeyValue("X-ENV", "dev"), entity.NewKeyValue("X-Team", "core"), entity.NewKeyValue("X-Folder", "f")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("headers = %+v, want %+v", got, want)
	}
	if collection.Headers[0].Value != "prod" {
		t.Fatal("ancestor headers must not be changed")
	}
}

func TestResolveEffectiveRequestCopiesTheRest(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "ws")
	request := entity.NewRequest("", "r")
	request.Body = entity.Body{Mode: bodymode.Binary, Binary: entity.FileRef{Src: "a.bin"}}
	request.Settings = entity.RequestSettings{TimeoutMs: 5, FollowRedirects: true}
	request.DisabledAutoHeaders = []string{"Accept"}
	effective := ResolveEffectiveRequest(request, nil, root)
	if effective.Body.Binary.Src != filepath.Join(root, "a.bin") || effective.Settings != request.Settings {
		t.Fatalf("body/settings = %+v %+v", effective.Body, effective.Settings)
	}
	effective.DisabledAutoHeaders[0] = "changed"
	if request.DisabledAutoHeaders[0] != "Accept" || !reflect.DeepEqual(ResolveEffectiveRequest(entity.NewRequest("", "r"), nil, root).DisabledAutoHeaders, []string{}) {
		t.Fatal("disabled auto headers must be copied")
	}
}

func autoHeaders(request entity.EffectiveRequest) map[string]entity.AutoHeader {
	result := map[string]entity.AutoHeader{}
	for _, header := range ComputeAutoHeaders(request, "1.2.3", "") {
		result[header.Key] = header
	}
	return result
}

func TestComputeAutoHeadersWithoutBody(t *testing.T) {
	got := ComputeAutoHeaders(entity.EffectiveRequest{URL: "https://api.test:8443/a"}, "1.2.3", "")
	want := []entity.AutoHeader{
		entity.NewComputedAutoHeader("Host", "api.test:8443"),
		entity.NewAutoHeader("User-Agent", "Xenon/1.2.3"),
		entity.NewAutoHeader("Accept", "*/*"),
		entity.NewAutoHeader("Accept-Encoding", "gzip, deflate, br"),
		entity.NewAutoHeader("Connection", "keep-alive"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ComputeAutoHeaders = %+v, want %+v", got, want)
	}
}

func TestComputeAutoHeadersWithBodyAndAuth(t *testing.T) {
	raw := autoHeaders(entity.EffectiveRequest{
		URL:  "{{host}}/a",
		Body: entity.Body{Mode: bodymode.Raw, Raw: entity.BodyRaw{Language: rawlanguage.Xml}},
		Auth: bearer("abc"),
	})
	if raw["Content-Type"] != entity.NewAutoHeader("Content-Type", "application/xml") {
		t.Errorf("raw content type = %+v", raw["Content-Type"])
	}
	if raw["Content-Length"] != entity.NewComputedAutoHeader("Content-Length", entity.ComputedAtSend) {
		t.Errorf("content length = %+v", raw["Content-Length"])
	}
	if raw["Host"] != entity.NewComputedAutoHeader("Host", entity.ComputedAtSend) {
		t.Errorf("unreadable host = %+v", raw["Host"])
	}
	if raw["Authorization"] != entity.NewAutoHeader("Authorization", "Bearer abc") {
		t.Errorf("authorization = %+v", raw["Authorization"])
	}
	multipart := autoHeaders(entity.EffectiveRequest{URL: "https://h", Body: entity.Body{Mode: bodymode.FormData}})
	if !multipart["Content-Type"].Computed {
		t.Errorf("multipart content type is computed at send: %+v", multipart["Content-Type"])
	}
}

func TestComputeAutoHeadersDisabledAndOverridden(t *testing.T) {
	headers := ComputeAutoHeaders(entity.EffectiveRequest{
		URL:                 "https://h",
		Auth:                bearer("abc"),
		Headers:             []entity.KeyValue{entity.NewKeyValue("accept", "application/json"), entity.NewKeyValue("AUTHORIZATION", "Token x")},
		DisabledAutoHeaders: []string{"user-agent"},
	}, "1", "")
	byKey := map[string]entity.AutoHeader{}
	for _, header := range headers {
		byKey[header.Key] = header
	}
	if _, found := byKey["Accept"]; found {
		t.Error("a user Accept header replaces the automatic one")
	}
	if _, found := byKey["Authorization"]; found {
		t.Error("a user Authorization header replaces the automatic one")
	}
	if !byKey["User-Agent"].Disabled || byKey["Connection"].Disabled || len(headers) != 4 {
		t.Errorf("headers = %+v", headers)
	}
}
