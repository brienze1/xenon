package entity

import (
	"reflect"
	"testing"
)

func orderedEnvironments() []Environment {
	return []Environment{
		{ID: "g", Name: "Gamma", FileName: "gamma.yaml"},
		{FileName: "zz-broken.yaml", Error: "invalid"},
		{ID: "a", Name: "alpha", FileName: "alpha.yaml"},
		{FileName: "aa-broken.yaml", Error: "invalid"},
		{ID: "b", Name: "Beta", FileName: "beta.yaml"},
		{ID: "d", Name: "Delta", FileName: "delta.yaml"},
	}
}

func fileNames(list EnvironmentList) []string {
	names := make([]string, 0, len(list.Environments))
	for _, environment := range list.Environments {
		names = append(names, environment.FileName)
	}
	return names
}

func TestNewEnvironmentListArrangesByOrderThenNameThenBrokenFiles(t *testing.T) {
	cases := []struct {
		order []string
		want  []string
	}{
		{order: nil, want: []string{"alpha.yaml", "beta.yaml", "delta.yaml", "gamma.yaml", "aa-broken.yaml", "zz-broken.yaml"}},
		{order: []string{"g", "b"}, want: []string{"gamma.yaml", "beta.yaml", "alpha.yaml", "delta.yaml", "aa-broken.yaml", "zz-broken.yaml"}},
		{order: []string{"gone", "d", "error:aa-broken.yaml", "a"}, want: []string{"delta.yaml", "alpha.yaml", "beta.yaml", "gamma.yaml", "aa-broken.yaml", "zz-broken.yaml"}},
		{order: []string{"b", "g", "b"}, want: []string{"beta.yaml", "gamma.yaml", "alpha.yaml", "delta.yaml", "aa-broken.yaml", "zz-broken.yaml"}},
		{order: []string{}, want: []string{"alpha.yaml", "beta.yaml", "delta.yaml", "gamma.yaml", "aa-broken.yaml", "zz-broken.yaml"}},
	}
	for _, tc := range cases {
		if got := fileNames(NewEnvironmentList(orderedEnvironments(), tc.order, "")); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("NewEnvironmentList(order %v) = %v, want %v", tc.order, got, tc.want)
		}
	}
}

func TestNewEnvironmentListKeepsOnlyAReadableActiveId(t *testing.T) {
	if got := NewEnvironmentList(orderedEnvironments(), nil, "b").ActiveID; got != "b" {
		t.Errorf("ActiveID = %q, want b", got)
	}
	if got := NewEnvironmentList(orderedEnvironments(), nil, "gone").ActiveID; got != "" {
		t.Errorf("ActiveID = %q, want empty", got)
	}
}

func TestEnvironmentListIDsSkipBrokenFiles(t *testing.T) {
	list := NewEnvironmentList(orderedEnvironments(), []string{"g"}, "")
	if got, want := list.IDs(), []string{"g", "a", "b", "d"}; !reflect.DeepEqual(got, want) {
		t.Errorf("IDs() = %v, want %v", got, want)
	}
	if !list.IsOrderedAs([]string{"g", "a", "b", "d"}) || list.IsOrderedAs([]string{"a", "b", "d", "g"}) {
		t.Errorf("IsOrderedAs mismatch for %v", list.IDs())
	}
}

func TestEnvironmentListOrderOperations(t *testing.T) {
	list := NewEnvironmentList(orderedEnvironments(), nil, "")
	cases := []struct {
		name string
		got  []string
		want []string
	}{
		{name: "move to the top", got: list.OrderMoving("g", 0), want: []string{"g", "a", "b", "d"}},
		{name: "move down", got: list.OrderMoving("a", 2), want: []string{"b", "d", "a", "g"}},
		{name: "move past the end", got: list.OrderMoving("a", 99), want: []string{"b", "d", "g", "a"}},
		{name: "move to the same place", got: list.OrderMoving("b", 1), want: []string{"a", "b", "d", "g"}},
		{name: "append", got: list.OrderAppending("n"), want: []string{"a", "b", "d", "g", "n"}},
		{name: "append a listed id", got: list.OrderAppending("b"), want: []string{"a", "b", "d", "g"}},
		{name: "replace in place", got: list.OrderReplacing("b", "n"), want: []string{"a", "n", "d", "g"}},
		{name: "replace an absent id", got: list.OrderReplacing("gone", "n"), want: []string{"a", "b", "d", "g", "n"}},
		{name: "import replacing", got: list.OrderAfterImport(ImportVariablesWrite{Environment: &Environment{ID: "n"}, Replaced: &Environment{ID: "b"}}), want: []string{"a", "n", "d", "g"}},
		{name: "import appending", got: list.OrderAfterImport(ImportVariablesWrite{Environment: &Environment{ID: "n"}}), want: []string{"a", "b", "d", "g", "n"}},
		{name: "remove", got: list.OrderRemoving("d"), want: []string{"a", "b", "g"}},
		{name: "remove an absent id", got: list.OrderRemoving("gone"), want: []string{"a", "b", "d", "g"}},
	}
	for _, tc := range cases {
		if !reflect.DeepEqual(tc.got, tc.want) {
			t.Errorf("%s = %v, want %v", tc.name, tc.got, tc.want)
		}
	}
	if got, want := list.IDs(), []string{"a", "b", "d", "g"}; !reflect.DeepEqual(got, want) {
		t.Errorf("the order operations changed the list: %v", got)
	}
}

func TestWorkspaceEnvironmentOrder(t *testing.T) {
	workspace := NewWorkspace("01W", "Acme")
	if workspace.EnvironmentOrder != nil || workspace.ListsEnvironment("a") {
		t.Fatalf("a new workspace has no environment order: %v", workspace.EnvironmentOrder)
	}
	ordered := workspace.WithEnvironmentOrder(nil)
	if ordered.EnvironmentOrder == nil || len(ordered.EnvironmentOrder) != 0 {
		t.Errorf("WithEnvironmentOrder(nil) = %v, want an empty stored order", ordered.EnvironmentOrder)
	}
	if listed := workspace.WithEnvironmentOrder([]string{"a"}); !listed.ListsEnvironment("a") || listed.ListsEnvironment("b") {
		t.Errorf("ListsEnvironment mismatch for %v", listed.EnvironmentOrder)
	}
}
