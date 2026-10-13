package entity

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

func TestAuthHeader(t *testing.T) {
	cases := []struct {
		auth  Auth
		value string
		found bool
	}{
		{Auth{Type: authtype.Bearer, Bearer: AuthBearer{Token: "abc"}}, "Bearer abc", true},
		{Auth{Type: authtype.Basic, Basic: AuthBasic{Username: "user", Password: "pass"}}, "Basic dXNlcjpwYXNz", true},
		{Auth{Type: authtype.Inherit, Bearer: AuthBearer{Token: "abc"}}, "", false},
		{Auth{Type: authtype.None}, "", false},
		{Auth{Type: authtype.ApiKey}, "", false},
	}
	for _, tc := range cases {
		if value, found := tc.auth.Header(); value != tc.value || found != tc.found {
			t.Errorf("%s: Header = %q %v", tc.auth.Type, value, found)
		}
	}
}

func TestBodyContentType(t *testing.T) {
	cases := []struct {
		body Body
		want string
	}{
		{Body{Mode: bodymode.None}, ""},
		{Body{Mode: bodymode.Raw, Raw: BodyRaw{Language: rawlanguage.Json}}, "application/json"},
		{Body{Mode: bodymode.Raw, Raw: BodyRaw{Language: rawlanguage.Xml}}, "application/xml"},
		{Body{Mode: bodymode.Raw, Raw: BodyRaw{Language: rawlanguage.Text}}, "text/plain"},
		{Body{Mode: bodymode.Raw, Raw: BodyRaw{Language: rawlanguage.Html}}, "text/html"},
		{Body{Mode: bodymode.Raw, Raw: BodyRaw{Language: rawlanguage.JavaScript}}, "application/javascript"},
		{Body{Mode: bodymode.Raw, Raw: BodyRaw{Language: rawlanguage.Unknown}}, "text/plain"},
		{Body{Mode: bodymode.UrlEncoded}, "application/x-www-form-urlencoded"},
		{Body{Mode: bodymode.FormData}, "multipart/form-data; boundary=<calculated when request is sent>"},
		{Body{Mode: bodymode.GraphQl}, "application/json"},
		{Body{Mode: bodymode.Binary}, "application/octet-stream"},
	}
	for _, tc := range cases {
		if got := tc.body.ContentType(); got != tc.want {
			t.Errorf("%s/%s: ContentType = %q, want %q", tc.body.Mode, tc.body.Raw.Language, got, tc.want)
		}
	}
	if !(Body{Mode: bodymode.FormData}).IsMultipart() || (Body{Mode: bodymode.Raw}).IsMultipart() {
		t.Error("only form-data is multipart")
	}
}

func TestBodyResolveFiles(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "ws")
	absolute := filepath.Join(string(filepath.Separator), "abs", "b.bin")
	body := Body{
		Mode:   bodymode.FormData,
		Binary: FileRef{Src: "files/a.bin"},
		FormData: []FormDataRow{
			{Key: "f", Type: formdatatype.File, Src: FileRef{Src: "files/c.txt"}, Enabled: true},
			{Key: "g", Type: formdatatype.File, Src: FileRef{Src: absolute}},
		},
	}
	resolved := body.ResolveFiles(root)
	if resolved.Binary.Src != filepath.Join(root, "files", "a.bin") {
		t.Fatalf("binary = %q", resolved.Binary.Src)
	}
	want := []FormDataRow{
		{Key: "f", Type: formdatatype.File, Src: FileRef{Src: filepath.Join(root, "files", "c.txt")}, Enabled: true},
		{Key: "g", Type: formdatatype.File, Src: FileRef{Src: absolute}},
	}
	if !reflect.DeepEqual(resolved.FormData, want) {
		t.Fatalf("form data = %+v", resolved.FormData)
	}
	if body.FormData[0].Src.Src != "files/c.txt" {
		t.Fatal("ResolveFiles must not change the receiver's rows")
	}
}

func TestRequestLogValueHidesSecrets(t *testing.T) {
	request := NewRequest("01J", "Me")
	request.URL = "https://h"
	request.Auth = Auth{Type: authtype.Bearer, Bearer: AuthBearer{Token: "secret"}}
	want := map[string]any{"id": "01J", "method": "GET", "url": "https://h"}
	if got := request.LogValue(); !reflect.DeepEqual(got, want) {
		t.Fatalf("LogValue = %v", got)
	}
}

func TestRequestSettingsTimeoutText(t *testing.T) {
	if got := (RequestSettings{TimeoutMs: 1500}).TimeoutText(); got != "1500" {
		t.Fatalf("TimeoutText = %q", got)
	}
}
