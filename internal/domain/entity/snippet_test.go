package entity

import (
	"reflect"
	"slices"
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

func TestNewSnippetRequestAddsTheAutoContentTypeExceptForMultipart(t *testing.T) {
	raw := Body{Mode: bodymode.Raw, Raw: BodyRaw{Language: rawlanguage.Json, Content: "{}"}}
	cases := []struct {
		request EffectiveRequest
		want    []KeyValue
	}{
		{request: EffectiveRequest{Body: raw, Headers: []KeyValue{NewKeyValue("A", "1")}}, want: []KeyValue{NewKeyValue("A", "1"), NewKeyValue(HeaderContentType, "application/json")}},
		{request: EffectiveRequest{Body: raw, Headers: []KeyValue{NewKeyValue("content-type", "text/x")}}, want: []KeyValue{NewKeyValue("content-type", "text/x")}},
		{request: EffectiveRequest{Body: raw, DisabledAutoHeaders: []string{"content-type"}}, want: []KeyValue{}},
		{request: EffectiveRequest{Body: Body{Mode: bodymode.FormData}}, want: []KeyValue{}},
		{request: EffectiveRequest{Body: Body{Mode: bodymode.None}}, want: []KeyValue{}},
	}
	for index, tc := range cases {
		got := NewSnippetRequest(tc.request, "/root", []string{"n"})
		if !reflect.DeepEqual(append([]KeyValue{}, got.Headers...), tc.want) || !slices.Equal(got.Notices, []string{"n"}) {
			t.Errorf("case %d: got %+v", index, got)
		}
	}
}

func TestSnippetRequestBodyState(t *testing.T) {
	cases := map[string]struct {
		body           Body
		present, empty bool
	}{
		"none":            {body: Body{Mode: bodymode.None}, empty: true},
		"raw":             {body: Body{Mode: bodymode.Raw, Raw: BodyRaw{Content: "x"}}, present: true},
		"empty raw":       {body: Body{Mode: bodymode.Raw}, present: true, empty: true},
		"urlencoded":      {body: Body{Mode: bodymode.UrlEncoded, UrlEncoded: []KeyValue{{Key: "a", Enabled: true}}}, present: true},
		"disabled rows":   {body: Body{Mode: bodymode.UrlEncoded, UrlEncoded: []KeyValue{{Key: "a"}}}, present: true, empty: true},
		"formdata":        {body: Body{Mode: bodymode.FormData, FormData: []FormDataRow{{Key: "a", Enabled: true}}}, present: true},
		"disabled form":   {body: Body{Mode: bodymode.FormData, FormData: []FormDataRow{{Key: "a"}}}, present: true, empty: true},
		"graphql":         {body: Body{Mode: bodymode.GraphQl, GraphQl: BodyGraphQl{Query: "q"}}, present: true},
		"empty graphql":   {body: Body{Mode: bodymode.GraphQl, GraphQl: BodyGraphQl{Variables: "{}"}}, present: true, empty: true},
		"binary":          {body: Body{Mode: bodymode.Binary, Binary: FileRef{Src: "a.bin"}}, present: true},
		"no binary file":  {body: Body{Mode: bodymode.Binary}, present: true, empty: true},
		"unset body mode": {body: Body{}, empty: true},
	}
	for name, tc := range cases {
		request := SnippetRequest{Body: tc.body}
		if request.HasBody() != tc.present || request.IsBodyEmpty() != tc.empty {
			t.Errorf("%s: HasBody %v IsBodyEmpty %v", name, request.HasBody(), request.IsBodyEmpty())
		}
	}
	request := SnippetRequest{Headers: []KeyValue{NewKeyValue("X-A", "1")}}
	if value, found := request.Header("x-a"); !found || value != "1" {
		t.Errorf("Header = %q %v", value, found)
	}
	if _, found := request.Header("x-b"); found {
		t.Error("unexpected header")
	}
}

func TestDerivedSecretsMaskCredentialsComputedFromSecrets(t *testing.T) {
	secrets := []string{"", "s3cret"}
	basic := Auth{Type: authtype.Basic, Basic: AuthBasic{Username: "ana", Password: "s3cret"}}
	aws := Auth{Type: authtype.AwsV4, AwsV4: AwsV4Auth{SecretKey: "s3cret"}}
	digest := Auth{Type: authtype.Digest, Digest: DigestAuth{Password: "s3cret"}}
	signed := EffectiveRequest{Headers: []KeyValue{NewKeyValue("X", "Signature=no"), NewKeyValue(HeaderAuthorization, "AWS4 Credential=a, Signature=abc")}}
	presigned := EffectiveRequest{URL: "https://x/?X-Amz-Signature=def"}
	answered := EffectiveRequest{Headers: []KeyValue{NewKeyValue(HeaderAuthorization, `Digest username="a", response="r1", qop=auth`)}}
	cases := []struct {
		auth    Auth
		request EffectiveRequest
		want    []string
	}{
		{auth: basic, want: []string{"YW5hOnMzY3JldA=="}},
		{auth: Auth{Type: authtype.Basic, Basic: AuthBasic{Username: "ana", Password: "plain"}}, want: []string{}},
		{auth: aws, request: signed, want: []string{"abc"}},
		{auth: aws, request: presigned, want: []string{"def"}},
		{auth: aws, want: []string{}},
		{auth: Auth{Type: authtype.AwsV4, AwsV4: AwsV4Auth{SecretKey: "plain"}}, request: signed, want: []string{}},
		{auth: digest, request: answered, want: []string{"r1"}},
		{auth: digest, want: []string{}},
		{auth: Auth{Type: authtype.Bearer, Bearer: AuthBearer{Token: "s3cret"}}, want: []string{}},
	}
	for index, tc := range cases {
		if got := tc.auth.DerivedSecrets(tc.request, secrets); !slices.Equal(got, tc.want) {
			t.Errorf("case %d: got %v, want %v", index, got, tc.want)
		}
	}
}

func TestSnippetSecretFormsCoverEncodingsAndEscapes(t *testing.T) {
	forms := SnippetSecretForms([]string{"", `a b'"$&`})
	for _, want := range []string{`a b'"$&`, `a+b%27%22%24%26`, `a%20b%27%22$&`, `a%20b%27%22%24%26`, `a b'\"$&`, `a b'\''"$&`, `a b\'"$&`, `a b'\"\$&`} {
		if !slices.Contains(forms, want) {
			t.Errorf("forms %q lack %q", forms, want)
		}
	}
}

func TestSnippetNoticesAndLogValues(t *testing.T) {
	mac := OAuth2Token{AccessToken: "t", TokenType: "mac"}
	bearer := OAuth2Token{AccessToken: "t", TokenType: "Bearer"}
	if got := OAuthSnippetNotices(nil); !slices.Equal(got, []string{NoOAuthTokenNotice}) {
		t.Errorf("nil token: %v", got)
	}
	if got := OAuthSnippetNotices(&OAuth2Token{}); !slices.Equal(got, []string{NoOAuthTokenNotice}) {
		t.Errorf("empty token: %v", got)
	}
	if got := OAuthSnippetNotices(&mac); !slices.Equal(got, []string{MacTokenNotice}) {
		t.Errorf("mac token: %v", got)
	}
	if got := OAuthSnippetNotices(&bearer); got != nil {
		t.Errorf("bearer token: %v", got)
	}
	if got := UnsupportedAuthNotice(Auth{Unsupported: AuthUnsupported{Type: "hawk"}}); got != "Auth type hawk is not supported, so the snippet has no auth." {
		t.Errorf("unsupported notice %q", got)
	}
	snippet := Snippet{Text: "abc", Notices: []string{"n"}}.WithText("abcd")
	if snippet.Text != "abcd" || !reflect.DeepEqual(snippet.LogValue(), map[string]any{"length": 4, "notices": []string{"n"}}) {
		t.Errorf("snippet %+v", snippet)
	}
	if value := (SnippetInput{LocalID: "l", IncludeSecrets: true}).LogValue().(map[string]any); value["localId"] != "l" || value["includeSecrets"] != true {
		t.Errorf("input log value %v", value)
	}
	if value := (SnippetRequest{Method: "GET", Headers: []KeyValue{{}}}).LogValue().(map[string]any); value["method"] != "GET" || value["headers"] != 1 {
		t.Errorf("request log value %v", value)
	}
}
