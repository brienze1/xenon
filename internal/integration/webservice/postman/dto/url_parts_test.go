package dto

import (
	"reflect"
	"testing"
)

func TestNewUrlPartsSplitsLikePostman(t *testing.T) {
	cases := []struct {
		raw  string
		want UrlParts
	}{
		{"{{base_url}}/users", UrlParts{Host: []string{"{{base_url}}"}, Path: []string{"users"}, HasPath: true}},
		{"https://api.acme.test:8443/v1/rebuilt?q=1#frag", UrlParts{Protocol: "https", Host: []string{"api", "acme", "test"}, Port: "8443", Path: []string{"v1", "rebuilt"}, HasPath: true, Hash: "frag"}},
		{"  http:///{{host}}.acme.test/a/:id/", UrlParts{Protocol: "http", Host: []string{"{{host}}", "acme", "test"}, Path: []string{"a", ":id", ""}, HasPath: true}},
		{"file:///tmp/x.json", UrlParts{Protocol: "file", Path: []string{"tmp", "x.json"}, HasPath: true}},
		{`{{url}}\path\to?x=1`, UrlParts{Host: []string{"{{url}}"}, Path: []string{"path", "to"}, HasPath: true}},
		{"https://user:pw@example.com", UrlParts{Protocol: "https", Host: []string{"example", "com"}}},
		{"http://[::1]:80/x", UrlParts{Protocol: "http", Host: []string{"[::1]"}, Port: "80", Path: []string{"x"}, HasPath: true}},
		{"http://[::1]/x", UrlParts{Protocol: "http", Host: []string{"[::1]"}, Path: []string{"x"}, HasPath: true}},
		{"{{a.b}}:{{port}}/{{p.q}}#{{h}}", UrlParts{Host: []string{"{{a.b}}"}, Port: "{{port}}", Path: []string{"{{p.q}}"}, HasPath: true, Hash: "{{h}}"}},
		{"localhost", UrlParts{Host: []string{"localhost"}}},
		{"", UrlParts{}},
	}
	for _, tc := range cases {
		if got := NewUrlParts(tc.raw); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("NewUrlParts(%q) = %+v, want %+v", tc.raw, got, tc.want)
		}
	}
}
