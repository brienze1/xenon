package dto

import (
	"errors"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestKeyValuesAcceptListsMapsAndNull(t *testing.T) {
	cases := map[string]struct {
		source string
		keys   []string
		values []string
	}{
		"list": {"rows:\n  - key: Accept\n    value: json\n  - key: X-Off\n    value: \"1\"\n    disabled: true\n", []string{"Accept", "X-Off"}, []string{"json", "1"}},
		"map":  {"rows:\n  Accept: json\n  X-Off: \"1\"\n", []string{"Accept", "X-Off"}, []string{"json", "1"}},
		"null": {"rows: null\n", []string{}, []string{}},
	}
	for name, tc := range cases {
		var document struct {
			Rows KeyValues `yaml:"rows"`
		}
		if err := yaml.Unmarshal([]byte(tc.source), &document); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(document.Rows) != len(tc.keys) {
			t.Fatalf("%s: got %d rows, want %d", name, len(document.Rows), len(tc.keys))
		}
		for index, row := range document.Rows {
			if row.Key != tc.keys[index] || text(row.Value) != tc.values[index] {
				t.Errorf("%s row %d: got %q=%q", name, index, row.Key, text(row.Value))
			}
		}
	}
}

func TestKeyValuesRefuseAScalarWithItsLine(t *testing.T) {
	var document struct {
		Rows KeyValues `yaml:"rows"`
	}
	err := yaml.Unmarshal([]byte("name: x\nrows: 's3cr3t'\n"), &document)
	var wrongType WrongTypeError
	if !errors.As(err, &wrongType) || wrongType.Line != 2 {
		t.Fatalf("got %v, want a wrong type error on line 2", err)
	}
}
