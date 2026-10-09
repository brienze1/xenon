package entity

import (
	"reflect"
	"testing"
)

func TestArrangedNamesFollowTheStoredOrderThenSortTheRest(t *testing.T) {
	cases := []struct {
		listed  []string
		present []string
		want    []string
	}{
		{listed: []string{"b", "gone", "a"}, present: []string{"a", "d", "b", "c"}, want: []string{"b", "a", "c", "d"}},
		{listed: []string{"a", "a"}, present: []string{"a"}, want: []string{"a"}},
		{listed: nil, present: []string{"z", "y"}, want: []string{"y", "z"}},
		{listed: []string{"x"}, present: []string{}, want: []string{}},
	}
	for _, tc := range cases {
		workspace := Workspace{Order: tc.listed}
		if got := workspace.ArrangedCollections(tc.present); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("ArrangedCollections(%v, %v) = %v, want %v", tc.listed, tc.present, got, tc.want)
		}
		collection := Collection{Order: tc.listed}
		if got := collection.ArrangedChildren(tc.present); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("ArrangedChildren(%v, %v) = %v, want %v", tc.listed, tc.present, got, tc.want)
		}
	}
}
