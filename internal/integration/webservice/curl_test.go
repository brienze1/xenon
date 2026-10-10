package webservice

import (
	"context"
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
	"github.com/brienze1/xenon/pkg/errs"
)

func readCurl(t *testing.T, text string) entity.ParsedCurl {
	t.Helper()
	parsed, err := NewCurlWebservice().ReadCommand(context.Background(), text)
	if err != nil {
		t.Fatalf("%s: %v", text, err)
	}
	return *parsed
}

func TestNormalizeTurnsEveryDialectIntoBash(t *testing.T) {
	cases := map[string][]string{
		"curl 'a b' \\\n  -H x":                            nil,
		"curl ^\"a b^\" ^\n  -H ^\"x^\"":                   {"curl", "a b", "-H", "x"},
		"curl ^\"x^\" \"a\"\"b\" \"c\\\"d\" e\\\\f \"\" g": {"curl", "x", `a"b`, `c"d`, `e\\f`, "", "g"},
		"curl ^\"x^\" \"a\\\\\\\"b\" ^^x":                  {"curl", "x", `a\"b`, "^x"},
		"curl ^\"x^\" \"a\\\\\" b":                         {"curl", "x", `a\`, "b"},
		"curl.exe 'it''s' `\n -d \"a`\"b\"\"c`n\"":         {"curl", "it's", "-d", "a\"b\"c\n"},
		"CURL.EXE `$x 'a b'":                               {"curl", "$x", "a b"},
		"curl -d 'x' `\n -H 'y'":                           {"curl", "-d", "x", "-H", "y"},
		"curl.exe \"t`tr`r\" ''":                           {"curl", "t\tr\r", ""},
	}
	for text, want := range cases {
		normalized := normalizeCurl(text)
		if want == nil {
			if normalized != text {
				t.Errorf("bash changed: %q", normalized)
			}
			continue
		}
		arguments, err := tokenizeCurl(normalized)
		if err != nil || !reflect.DeepEqual(append([]string{"curl"}, arguments...), want) {
			t.Errorf("%q: got %q (%q) %v, want %q", text, arguments, normalized, err, want)
		}
	}
	if got := normalizeCurl("curl a\r\nb\rc"); got != "curl a\nb\nc" {
		t.Errorf("line endings = %q", got)
	}
}

func TestTokenizeDecodesShellQuoting(t *testing.T) {
	cases := map[string][]string{
		`curl a\ b "c\$d\"\\\x" 'e\f' $'g\n\t\x41é\U0001F600\q\'\"\\' x\`: {"a b", "c$d\"\\\\x", `e\f`, "g\n\tAé😀\\q'\"\\", "x\\"},
		"curl a\\\nb \"c\\\nd\"":                 {"ab", "cd"},
		`curl $HOME "${x}y" $(z) $'\x4' $'\u12'`: {"$HOME", "${x}y", "$(z)", `\x4`, `\u12`},
		`curl a; curl b`:                         {"a"},
		`x=1 curl a`:                             {"a"},
		`CURL a`:                                 {"a"},
	}
	for text, want := range cases {
		arguments, err := tokenizeCurl(text)
		if err != nil || !reflect.DeepEqual(arguments, want) {
			t.Errorf("%q: got %q %v, want %q", text, arguments, err, want)
		}
	}
	refused := map[string]string{
		"curl 'a":             "The cURL command could not be read",
		"wget a":              "The text is not a cURL command",
		"":                    "The text is not a cURL command",
		"if true; then a; fi": "The text is not a cURL command",
	}
	for text, message := range refused {
		_, err := tokenizeCurl(text)
		if custom, isCustom := err.(errs.BaseError); !isCustom || custom.Code() != "IMP-06400" || custom.Description() != message {
			t.Errorf("%q: got %v", text, err)
		}
	}
}

func TestReadCommandAppliesTheFlagTable(t *testing.T) {
	parsed := readCurl(t, `curl -sSLk -XPOST -H'A: 1' --header 'B:2' -H 'C' -b 'a=1' --cookie 'b=2' -A ua -e ref --user u:p -m 5 -o out --retry 3 -Z --frobnicate -- u1`)
	request := parsed.Request
	if request.Method != "POST" || request.URL != "u1" || request.Settings.VerifySsl || !request.Settings.FollowRedirects {
		t.Fatalf("request = %+v", request)
	}
	wantHeaders := []entity.KeyValue{
		entity.NewKeyValue("A", "1"), entity.NewKeyValue("B", "2"), entity.NewKeyValue("C", ""),
		entity.NewKeyValue("Cookie", "a=1; b=2"), entity.NewKeyValue("User-Agent", "ua"), entity.NewKeyValue("Referer", "ref"),
	}
	if !reflect.DeepEqual(request.Headers, wantHeaders) {
		t.Errorf("headers = %+v", request.Headers)
	}
	if request.Auth.Type != authtype.Basic || request.Auth.Basic.Username != "u" || request.Auth.Basic.Password != "p" {
		t.Errorf("auth = %+v", request.Auth)
	}
	if !reflect.DeepEqual(parsed.Notices, []string{"Ignored unknown option -Z", "Ignored unknown option --frobnicate", "Ignored unknown option --"}) {
		t.Errorf("notices = %q", parsed.Notices)
	}
}

func TestReadCommandBuildsBodies(t *testing.T) {
	cases := map[string]struct {
		text     string
		mode     string
		content  string
		language string
		method   string
	}{
		"form-like data":        {`curl u -d 'a=1&b=two%20words'`, bodymode.UrlEncoded.String(), "", "", "POST"},
		"text data":             {`curl u --data-ascii 'plain'`, bodymode.Raw.String(), "plain", rawlanguage.Text.String(), "POST"},
		"json data":             {`curl u -d '{"a":1}' -d '[2]'`, bodymode.Raw.String(), `{"a":1}&[2]`, rawlanguage.Text.String(), "POST"},
		"json detected":         {`curl u --data-raw ' [1]'`, bodymode.Raw.String(), " [1]", rawlanguage.Json.String(), "POST"},
		"xml type":              {`curl u -H 'content-type: text/xml' -d 'a=1'`, bodymode.Raw.String(), "a=1", rawlanguage.Xml.String(), "POST"},
		"html type":             {`curl u -H 'Content-Type: text/html' -d x`, bodymode.Raw.String(), "x", rawlanguage.Html.String(), "POST"},
		"javascript type":       {`curl u -H 'Content-Type: application/javascript' -d x`, bodymode.Raw.String(), "x", rawlanguage.JavaScript.String(), "POST"},
		"form type":             {`curl u -H 'Content-Type: application/x-www-form-urlencoded; charset=utf-8' -d 'a b'`, bodymode.UrlEncoded.String(), "", "", "POST"},
		"json flag":             {`curl u --json '{"a":' --json '1}' -H 'accept: x'`, bodymode.Raw.String(), `{"a":1}`, rawlanguage.Json.String(), "POST"},
		"raw at sign":           {`curl u --data-raw '@x'`, bodymode.Raw.String(), "@x", rawlanguage.Text.String(), "POST"},
		"binary file":           {`curl u --data-binary @f -X PATCH`, bodymode.Binary.String(), "", "", "PATCH"},
		"upload":                {`curl u -T f`, bodymode.Binary.String(), "", "", "PUT"},
		"head":                  {`curl -I u`, bodymode.None.String(), "", "", "HEAD"},
		"get":                   {`curl -G u`, bodymode.None.String(), "", "", "GET"},
		"form":                  {`curl u -F 'a=1' --form-string 'b=@x' -F 'c=<f;type=t'`, bodymode.FormData.String(), "", "", "POST"},
		"urlencode":             {`curl u --data-urlencode 'q=a b' --data-urlencode 'plain' -d 'c=1'`, bodymode.UrlEncoded.String(), "", "", "POST"},
		"empty key not formish": {`curl u -d '=1'`, bodymode.Raw.String(), "=1", rawlanguage.Text.String(), "POST"},
	}
	for name, tc := range cases {
		request := readCurl(t, tc.text).Request
		if request.Body.Mode.String() != tc.mode || request.Method != tc.method {
			t.Errorf("%s: mode %s method %s", name, request.Body.Mode, request.Method)
			continue
		}
		if tc.content != "" && (request.Body.Raw.Content != tc.content || request.Body.Raw.Language.String() != tc.language) {
			t.Errorf("%s: raw = %+v", name, request.Body.Raw)
		}
	}
	form := readCurl(t, `curl u -F 'a=1' -F 'b=@/x.pdf;type=application/pdf' -F 'c=<f'`).Request.Body.FormData
	if len(form) != 3 || form[0].Value != "1" || form[1].Src.Src != "/x.pdf" || !form[1].Src.NeedsRepick() || !form[1].IsFile() || form[2].Src.Src != "f" {
		t.Errorf("form = %+v", form)
	}
	encoded := readCurl(t, `curl u --data-urlencode 'q=a b' --data-urlencode 'plain' -d 'c=1%2B2&d'`).Request.Body.UrlEncoded
	want := []entity.KeyValue{entity.NewKeyValue("q", "a b"), entity.NewKeyValue("", "plain"), entity.NewKeyValue("c", "1+2"), entity.NewKeyValue("d", "")}
	if !reflect.DeepEqual(encoded, want) {
		t.Errorf("urlencoded = %+v", encoded)
	}
	if bad := readCurl(t, `curl u -d 'a=%zz' --data-urlencode x=y`).Request.Body.UrlEncoded; bad[0].Value != "%zz" {
		t.Errorf("bad escape = %+v", bad)
	}
	headers := readCurl(t, `curl u --json 1 -H 'accept: x'`).Request.Headers
	if !reflect.DeepEqual(headers, []entity.KeyValue{entity.NewKeyValue("accept", "x"), entity.NewKeyValue("Content-Type", "application/json")}) {
		t.Errorf("json headers = %+v", headers)
	}
}

func TestReadCommandAppendsGetDataToTheUrl(t *testing.T) {
	cases := map[string]string{
		`curl -G 'u?x=1' -d 'a=1' --data-urlencode 'q=a b' --data-urlencode 'only space'`: "u?x=1&a=1&q=a%20b&only%20space",
		`curl -G u -d a=1`: "u?a=1",
		`curl -G u`:        "u",
	}
	for text, want := range cases {
		request := readCurl(t, text).Request
		if request.URL != want || request.Body.Mode != bodymode.None || request.Method != "GET" {
			t.Errorf("%s: url %q mode %s", text, request.URL, request.Body.Mode)
		}
	}
}

func TestReadCommandMapsAuth(t *testing.T) {
	digest := readCurl(t, `curl --digest -u 'a:b:c' u`).Request.Auth
	if digest.Type != authtype.Digest || !reflect.DeepEqual(digest.Attributes, []entity.AuthAttribute{{Key: "username", Value: "a"}, {Key: "password", Value: "b:c"}}) {
		t.Errorf("digest = %+v", digest)
	}
	basic := readCurl(t, `curl --digest --basic -u a u`).Request.Auth
	if basic.Type != authtype.Basic || basic.Basic.Username != "a" || basic.Basic.Password != "" {
		t.Errorf("basic = %+v", basic)
	}
	if none := readCurl(t, `curl u`).Request.Auth; none.Type != authtype.Inherit {
		t.Errorf("none = %+v", none)
	}
}

func TestReadCommandPicksTheHttpUrl(t *testing.T) {
	parsed := readCurl(t, `curl ftp://a HTTPS://b --url http://c u -b cookies.txt -x`)
	if parsed.Request.URL != "HTTPS://b" {
		t.Errorf("url = %q", parsed.Request.URL)
	}
	want := []string{"Ignored the cookie file cookies.txt", "Ignored the extra URL ftp://a", "Ignored the extra URL http://c", "Ignored the extra URL u"}
	if !reflect.DeepEqual(parsed.Notices, want) {
		t.Errorf("notices = %q", parsed.Notices)
	}
	refused := map[string]string{
		"curl -X POST":        "IMP-06400",
		"curl -H":             "IMP-06400",
		"curl ws://a ftp://b": "IMP-07400",
		"curl 'a":             "IMP-06400",
	}
	for text, code := range refused {
		_, err := NewCurlWebservice().ReadCommand(context.Background(), text)
		if custom, isCustom := err.(errs.BaseError); !isCustom || custom.Code() != code {
			t.Errorf("%q: got %v", text, err)
		}
	}
	_, err := NewCurlWebservice().ReadCommand(context.Background(), "curl ws://a ftp://b")
	if custom, _ := err.(errs.BaseError); custom.Description() != "Only HTTP and HTTPS URLs can be imported, not ws://a" {
		t.Errorf("message = %v", err)
	}
}
