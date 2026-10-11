package enums_test

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/clientauthentication"
	"github.com/brienze1/xenon/internal/domain/enums/digestalgorithm"
	"github.com/brienze1/xenon/internal/domain/enums/digestqop"
	"github.com/brienze1/xenon/internal/domain/enums/dropposition"
	"github.com/brienze1/xenon/internal/domain/enums/oauth2grant"
	"github.com/brienze1/xenon/internal/domain/enums/snippetlanguage"
	"github.com/brienze1/xenon/internal/domain/enums/tokenplacement"
)

func TestStageFiveEnumsParseTheirValues(t *testing.T) {
	cases := []struct {
		got  string
		want string
	}{
		{oauth2grant.GetOAuth2Grant(" Authorization_Code ").String(), "authorization_code"},
		{oauth2grant.GetOAuth2Grant("authorization_code_with_pkce").String(), "authorization_code_with_pkce"},
		{oauth2grant.GetOAuth2Grant("implicit").String(), "implicit"},
		{oauth2grant.GetOAuth2Grant("password_credentials").String(), "password_credentials"},
		{oauth2grant.GetOAuth2Grant("client_credentials").String(), "client_credentials"},
		{oauth2grant.GetOAuth2Grant("device").String(), "unknown"},
		{clientauthentication.GetClientAuthentication("HEADER").String(), "header"},
		{clientauthentication.GetClientAuthentication("body").String(), "body"},
		{clientauthentication.GetClientAuthentication("basic").String(), "unknown"},
		{tokenplacement.GetTokenPlacement("header").String(), "header"},
		{tokenplacement.GetTokenPlacement("queryparams").String(), "queryParams"},
		{tokenplacement.GetTokenPlacement("cookie").String(), "unknown"},
		{digestalgorithm.GetDigestAlgorithm("md5").String(), "MD5"},
		{digestalgorithm.GetDigestAlgorithm("MD5-SESS").String(), "MD5-sess"},
		{digestalgorithm.GetDigestAlgorithm("sha-256").String(), "SHA-256"},
		{digestalgorithm.GetDigestAlgorithm("SHA-256-sess").String(), "SHA-256-sess"},
		{digestalgorithm.GetDigestAlgorithm("SHA-512-256").String(), "SHA-512-256"},
		{digestalgorithm.GetDigestAlgorithm("SHA-512-256-sess").String(), "SHA-512-256-sess"},
		{digestalgorithm.GetDigestAlgorithm("SHA-1").String(), "unknown"},
		{digestqop.GetDigestQop("auth").String(), "auth"},
		{digestqop.GetDigestQop("AUTH-INT").String(), "auth-int"},
		{digestqop.GetDigestQop(" ").String(), ""},
		{digestqop.GetDigestQop("other").String(), "unknown"},
		{snippetlanguage.GetSnippetLanguage("curl").String(), "curl"},
		{snippetlanguage.GetSnippetLanguage("js_fetch").String(), "js_fetch"},
		{snippetlanguage.GetSnippetLanguage("go_native").String(), "go_native"},
		{snippetlanguage.GetSnippetLanguage("python_requests").String(), "python_requests"},
		{snippetlanguage.GetSnippetLanguage("ruby").String(), "unknown"},
		{dropposition.GetDropPosition("before").String(), "before"},
		{dropposition.GetDropPosition("Inside").String(), "inside"},
		{dropposition.GetDropPosition("after").String(), "after"},
		{dropposition.GetDropPosition("below").String(), "unknown"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}

func TestStageFiveEnumsCompareIgnoringCase(t *testing.T) {
	checks := map[string]bool{
		"grant":     oauth2grant.Implicit.Equals("IMPLICIT"),
		"client":    clientauthentication.Body.Equals("Body"),
		"placement": tokenplacement.QueryParams.Equals("queryparams"),
		"algorithm": digestalgorithm.MD5Sess.Equals("md5-SESS"),
		"qop":       digestqop.AuthInt.Equals("Auth-Int"),
		"language":  snippetlanguage.JsFetch.Equals("JS_FETCH"),
		"drop":      dropposition.Inside.Equals("INSIDE"),
	}
	for name, equal := range checks {
		if !equal {
			t.Errorf("%s: Equals must ignore case", name)
		}
	}
	if dropposition.Before.Equals("after") || oauth2grant.Implicit.Equals("implicit2") {
		t.Error("different values must not be equal")
	}
}
