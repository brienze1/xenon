package dto

import (
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func TestDecodeErrorsNeverQuoteTheContent(t *testing.T) {
	cases := map[string]struct {
		document entity.ImportDocument
		target   any
		want     string
	}{
		"yaml type":   {entity.ImportDocument{Path: "a.yaml", Data: []byte("name: x\nvalues: 's3cr3t-unit'\n")}, &Environment{}, "a.yaml is not a valid Postman v3 file: line 2: a value has the wrong type"},
		"rows type":   {entity.ImportDocument{Path: "b.request.yaml", Data: []byte("$kind: http-request\nheaders: 's3cr3t-unit'\n")}, &Request{}, "b.request.yaml is not a valid Postman v3 file: line 2: a value has the wrong type"},
		"json type":   {entity.ImportDocument{Path: "j.json", Data: []byte(`{"name": "x", "values": "s3cr3t-unit"}`)}, &Environment{}, "j.json is not a valid Postman v3 file: a value has the wrong type"},
		"json syntax": {entity.ImportDocument{Path: "c.json", Data: []byte(`{"name": "s3cr3t-unit",`)}, &Environment{}, "c.json is not a valid Postman v3 file: byte 23: invalid JSON"},
		"yaml syntax": {entity.ImportDocument{Path: "d.yaml", Data: []byte("name: [s3cr3t-unit\n")}, &Environment{}, "d.yaml is not a valid Postman v3 file: line 1: did not find expected ',' or ']'"},
	}
	for name, tc := range cases {
		err := Decode(tc.document, tc.target)
		if err == nil || err.Error() != tc.want || strings.Contains(err.Error(), "s3cr3t-unit") {
			t.Errorf("%s: got %v, want %q", name, err, tc.want)
		}
	}
	if got := NewDecodeError("e.request.yaml", UnknownKindError{Kind: "x"}).Error(); got != `e.request.yaml is not a valid Postman v3 file: unknown $kind "x"` {
		t.Errorf("unknown kind: got %q", got)
	}
}
