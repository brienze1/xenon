package entity

import (
	"reflect"
	"testing"
)

func TestTreeSearchCarriesTheTreeItSearched(t *testing.T) {
	search := NewTreeSearch()
	search.AddMatch("r1", nil)
	tree := sampleTree()
	got := search.WithTree(tree)
	if !reflect.DeepEqual(got.Tree, tree) || !reflect.DeepEqual(got.MatchIDs, []string{"r1"}) {
		t.Fatalf("WithTree = %+v, want the matches and the searched tree", got)
	}
	if search.Tree.Collections() != nil {
		t.Fatalf("WithTree changed the original search: %+v", search)
	}
}
