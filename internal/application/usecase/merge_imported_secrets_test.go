package usecase

import (
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
)

func TestMergeImportedSecrets(t *testing.T) {
	existing := []entity.VariableRow{{Key: "shared", Secret: true}, {Key: "plain"}}
	imported := []entity.VariableRow{{Key: "shared", Secret: true}, {Key: "added", Secret: true}, {Key: "plain"}, {Key: "unset", Secret: true}}
	current := map[string]string{"shared": "old", "plain": "stray", "kept": "mine"}
	values := map[string]string{"shared": "new", "added": "fresh", "plain": "ignored"}
	cases := []struct {
		resolution enums.ClashResolution
		expected   map[string]string
	}{
		{clashresolution.Replace, map[string]string{"shared": "new", "added": "fresh", "kept": "mine"}},
		{clashresolution.KeepBoth, map[string]string{"shared": "old", "plain": "stray", "added": "fresh", "kept": "mine"}},
		{clashresolution.Unresolved, map[string]string{"shared": "old", "plain": "stray", "added": "fresh", "kept": "mine"}},
	}
	for _, c := range cases {
		got := MergeImportedSecrets(existing, imported, current, values, c.resolution)
		if !reflect.DeepEqual(got, c.expected) {
			t.Errorf("%s: got %v, expected %v", c.resolution, got, c.expected)
		}
	}
	if current["shared"] != "old" || len(current) != 3 {
		t.Errorf("current was changed: %v", current)
	}
}
