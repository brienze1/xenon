package usecase

import (
	"encoding/json"
	"testing"

	"github.com/brienze1/xenon/pkg/errs"
)

func TestParseIterationDataReadsACSVWithItsHeaderRow(t *testing.T) {
	data, err := ParseIterationData("/data/users.csv", []byte("\xEF\xBB\xBFemail,password\r\na@x.test,pa,extra\nb@x.test\n"))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if data.FileName != "users.csv" || data.Count() != 2 {
		t.Fatalf("data = %+v", data)
	}
	if data.Rows[0]["email"] != "a@x.test" || data.Rows[0]["password"] != "pa" || len(data.Rows[0]) != 2 {
		t.Errorf("extra cells are dropped: %v", data.Rows[0])
	}
	if data.Rows[1]["password"] != "" {
		t.Errorf("missing cells are empty: %v", data.Rows[1])
	}
}

func TestParseIterationDataKeepsJSONTypes(t *testing.T) {
	data, err := ParseIterationData("/data/rows.json", []byte(`[{ "n": 1.5, "ok": true, "none": null, "obj": { "a": 1 } }]`))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	row := data.Rows[0]
	if row["n"] != json.Number("1.5") || row["ok"] != true || row["none"] != nil {
		t.Errorf("row = %v", row)
	}
	if object, isObject := row["obj"].(map[string]any); !isObject || object["a"] != json.Number("1") {
		t.Errorf("obj = %v", row["obj"])
	}
}

func TestParseIterationDataDetectsJSONWithoutTheExtension(t *testing.T) {
	data, err := ParseIterationData("/data/rows.txt", []byte(` [{"a":"b"}]`))
	if err != nil || data.Rows[0]["a"] != "b" {
		t.Errorf("data = %+v, err = %v", data, err)
	}
}

func TestParseIterationDataRefusesUnusableFiles(t *testing.T) {
	cases := map[string]struct {
		path    string
		content string
	}{
		"empty CSV":            {path: "empty.csv", content: ""},
		"header only":          {path: "header.csv", content: "email\n"},
		"broken CSV":           {path: "broken.csv", content: "a,b\n\"x\"y\"z,\"\n\"unterminated"},
		"JSON object":          {path: "obj.json", content: `{"email":"a"}`},
		"JSON array of values": {path: "vals.json", content: `[1, 2]`},
		"empty JSON array":     {path: "none.json", content: `[]`},
		"invalid JSON":         {path: "bad.json", content: `[{`},
		"not UTF-8":            {path: "latin.csv", content: "name\n\xff\xfe\n"},
	}
	for name, test := range cases {
		_, err := ParseIterationData(test.path, []byte(test.content))
		failure, isBase := err.(errs.BaseError)
		if !isBase || failure.Code() != "RUN-05400" {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
