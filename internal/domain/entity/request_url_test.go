package entity

import (
	"reflect"
	"testing"
)

func TestRequestUrlStringRoundTrips(t *testing.T) {
	for _, raw := range []string{"", "https://h/a", "https://h/a?", "https://h/a?x=1", "https://h/a#top", "https://h/a?x=1#top", "https://h/a#x?y"} {
		if got := ParseRequestUrl(raw).String(); got != raw {
			t.Errorf("ParseRequestUrl(%q).String() = %q", raw, got)
		}
	}
}

func TestRequestUrlQueryParams(t *testing.T) {
	got := ParseRequestUrl("https://h/a?a=1&&flag&b=x=y&c=#frag=1").QueryParams()
	want := []KeyValue{NewKeyValue("a", "1"), NewKeyValue("flag", ""), NewKeyValue("b", "x=y"), NewKeyValue("c", "")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("QueryParams = %+v, want %+v", got, want)
	}
	if params := ParseRequestUrl("https://h/a").QueryParams(); len(params) != 0 || params == nil {
		t.Fatalf("no query must give an empty list, got %#v", params)
	}
}

func TestRequestUrlWithQuery(t *testing.T) {
	rows := []KeyValue{
		NewKeyValue("a", "1"),
		{Key: "off", Value: "1"},
		NewKeyValue("flag", ""),
		NewKeyValue("  ", "blank key"),
		NewKeyValue("q", "a b"),
	}
	if got := ParseRequestUrl("https://h/a?old=1#top").WithQuery(rows).String(); got != "https://h/a?a=1&flag&q=a b#top" {
		t.Fatalf("WithQuery = %q", got)
	}
	if got := ParseRequestUrl("https://h/a?old=1").WithQuery([]KeyValue{{Key: "off"}}).String(); got != "https://h/a" {
		t.Fatalf("no active rows must drop the question mark, got %q", got)
	}
}

func TestRequestUrlPathVariableNames(t *testing.T) {
	got := ParseRequestUrl("http://localhost:8080/{{base}}/users/:id/copies/:id/a:b/:/:postId?x=:y#:z").PathVariableNames()
	if !reflect.DeepEqual(got, []string{"id", "postId"}) {
		t.Fatalf("PathVariableNames = %v", got)
	}
	if names := ParseRequestUrl("https://h/a").PathVariableNames(); names == nil || len(names) != 0 {
		t.Fatalf("no variables must give an empty list, got %#v", names)
	}
}

func TestRequestUrlWithPathVariableNames(t *testing.T) {
	cases := map[string]struct {
		url   string
		names []string
		want  string
	}{
		"renames by position":         {"https://h/users/:id/posts/:postId?x=1", []string{"userId", "postId"}, "https://h/users/:userId/posts/:postId?x=1"},
		"renames every repeat":        {"https://h/:id/x/:id", []string{"key"}, "https://h/:key/x/:key"},
		"plain segments untouched":    {"https://h/id/:id", []string{"userId"}, "https://h/id/:userId"},
		"blank names keep the old":    {"https://h/:a/:b", []string{" ", "c"}, "https://h/:a/:c"},
		"fewer names keep the rest":   {"https://h/:a/:b", []string{"c"}, "https://h/:c/:b"},
		"more names are ignored":      {"https://h/:a", []string{"b", "c"}, "https://h/:b"},
		"trimmed replacement":         {"https://h/:a", []string{" b "}, "https://h/:b"},
		"no variables nothing to do":  {"https://h/a", []string{"b"}, "https://h/a"},
		"empty list keeps everything": {"https://h/:a", nil, "https://h/:a"},
	}
	for name, tc := range cases {
		if got := ParseRequestUrl(tc.url).WithPathVariableNames(tc.names).String(); got != tc.want {
			t.Errorf("%s: got %q, want %q", name, got, tc.want)
		}
	}
}

func TestRequestUrlWithPathValues(t *testing.T) {
	variables := []PathVariable{{Key: "id", Value: "42"}, {Key: "empty"}, {Key: "postId", Value: "7"}}
	got := ParseRequestUrl("https://h/id/:id/:empty/:postId/:other?q=:id").WithPathValues(variables).String()
	if got != "https://h/id/42/:empty/7/:other?q=:id" {
		t.Fatalf("WithPathValues = %q", got)
	}
}

func TestRequestUrlWithDefaultScheme(t *testing.T) {
	cases := map[string]string{
		"":                   "",
		"   ":                "   ",
		"api.test/a":         "http://api.test/a",
		"https://api.test/a": "https://api.test/a",
		"ftp://x":            "ftp://x",
	}
	for raw, want := range cases {
		if got := ParseRequestUrl(raw).WithDefaultScheme().String(); got != want {
			t.Errorf("WithDefaultScheme(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestRequestUrlWithoutFragment(t *testing.T) {
	if got := ParseRequestUrl("https://h/a?x=1#top").WithoutFragment().String(); got != "https://h/a?x=1" {
		t.Fatalf("WithoutFragment = %q", got)
	}
}

func TestRequestUrlEncoded(t *testing.T) {
	cases := map[string]string{
		"https://h/files/my file.txt?q=a|b c&done=a%20b&json={\"a\":1}": "https://h/files/my%20file.txt?q=a%7Cb%20c&done=a%20b&json=%7B%22a%22:1%7D",
		"https://h":                        "https://h",
		"https://h/":                       "https://h/",
		"https://user@h:8080/a@b:c/~x_y":   "https://user@h:8080/a@b:c/~x_y",
		"https://h/café":                   "https://h/caf%C3%A9",
		"https://h/a%2fb%zz%4":             "https://h/a%2fb%25zz%254",
		"https://h/a%":                     "https://h/a%25",
		"https://h/a?p=/x?y:z@!$&'()*+,;=": "https://h/a?p=/x?y:z@!$&'()*+,;=",
		"https://h/a?b=[x]^`\\":            "https://h/a?b=%5Bx%5D%5E%60%5C",
		"https://h/AZaz09-._~":             "https://h/AZaz09-._~",
		"no scheme/a b?c d":                "no scheme/a b?c%20d",
		"https://h/a?x=%41%4a%4G":          "https://h/a?x=%41%4a%254G",
		"https://h/a?x=%G1":                "https://h/a?x=%25G1",
	}
	for raw, want := range cases {
		if got := ParseRequestUrl(raw).Encoded().String(); got != want {
			t.Errorf("Encoded(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestRequestUrlHost(t *testing.T) {
	cases := map[string]string{
		"https://api.test:8443/users": "api.test:8443",
		"api.test/users":              "api.test",
		"{{host}}/users":              "",
		"":                            "",
	}
	for raw, want := range cases {
		if got := ParseRequestUrl(raw).Host(); got != want {
			t.Errorf("Host(%q) = %q, want %q", raw, got, want)
		}
	}
}
