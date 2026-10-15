package entity

import (
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/importkind"
)

func TestNewImportSourceReportsTheSharedKindOrMixed(t *testing.T) {
	collection21 := ImportInput{Kind: importkind.PostmanCollection, Version: "2.1"}
	collection20 := ImportInput{Kind: importkind.PostmanCollection, Version: "2.0"}
	globals := ImportInput{Kind: importkind.PostmanGlobals}
	cases := map[string]struct {
		inputs  []ImportInput
		kind    enums.ImportKind
		version string
	}{
		"same kind and version":         {[]ImportInput{collection21, collection21}, importkind.PostmanCollection, "2.1"},
		"same kind, different versions": {[]ImportInput{collection21, collection20}, importkind.Mixed, ""},
		"different kinds":               {[]ImportInput{globals, collection21}, importkind.Mixed, ""},
		"one input":                     {[]ImportInput{collection20}, importkind.PostmanCollection, "2.0"},
		"no inputs":                     {nil, "", ""},
	}
	for name, tc := range cases {
		source := NewImportSource("s", tc.inputs, []string{}, time.Unix(0, 0))
		if source.Kind != tc.kind || source.Version != tc.version {
			t.Errorf("%s: got %q %q, want %q %q", name, source.Kind, source.Version, tc.kind, tc.version)
		}
	}
}
