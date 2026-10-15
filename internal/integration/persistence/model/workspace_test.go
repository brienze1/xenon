package model

import (
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

func patched(t *testing.T, source string, fields []yamlnode.Field) string {
	t.Helper()
	document, err := yamlnode.Parse([]byte(source))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	yamlnode.Patch(yamlnode.Root(document), fields)
	out, err := yamlnode.Encode(document)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	return string(out)
}

func TestWorkspaceWithoutEnvironmentOrderWritesNoEnvironmentsField(t *testing.T) {
	var stored Workspace
	read(t, "name: Acme\norder: []\n", &stored)
	if order := stored.ToEntity().EnvironmentOrder; order != nil {
		t.Fatalf("EnvironmentOrder = %v, want nil", order)
	}
	model := NewWorkspace(stored.ToEntity())
	source := "name: Acme\norder: []\nenvironments: oops\nvariables: []\n"
	if got := patched(t, source, model.Fields()); got != "xenon: 1\nkind: workspace\nid: \"\"\nname: Acme\ndescription: \"\"\norder: []\nenvironments: oops\nvariables: []\nsync: {}\n" {
		t.Fatalf("got:\n%s", got)
	}
}

func TestWorkspaceMalformedEnvironmentsReadAsNoOrder(t *testing.T) {
	for _, source := range []string{
		"environments: oops\n",
		"environments: [a, b]\n",
		"environments:\n  order: oops\n",
		"environments:\n  order:\n    - a\n    - [b]\n",
		"environments:\n  other: 1\n",
	} {
		var stored Workspace
		read(t, source, &stored)
		if order := stored.ToEntity().EnvironmentOrder; order != nil {
			t.Errorf("%q: EnvironmentOrder = %v, want nil", source, order)
		}
	}
}

func TestWorkspaceEnvironmentOrderRoundTrips(t *testing.T) {
	workspace := entity.NewWorkspace("01W", "Acme").WithEnvironmentOrder([]string{"01P", "01D"})
	model := NewWorkspace(workspace)
	want := "xenon: 1\nkind: workspace\nid: 01W\nname: Acme\ndescription: \"\"\norder: []\nenvironments:\n  order:\n    - 01P\n    - 01D\nvariables: []\nsync: {}\n"
	got := write(t, model.Fields())
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	var decoded Workspace
	read(t, got, &decoded)
	if !reflect.DeepEqual(decoded.ToEntity(), workspace) {
		t.Fatalf("decoded = %+v, want %+v", decoded.ToEntity(), workspace)
	}
}

func TestWorkspaceEmptyEnvironmentOrderIsWritten(t *testing.T) {
	model := NewWorkspace(entity.NewWorkspace("01W", "Acme").WithEnvironmentOrder(nil))
	got := write(t, model.Fields())
	if want := "xenon: 1\nkind: workspace\nid: 01W\nname: Acme\ndescription: \"\"\norder: []\nenvironments:\n  order: []\nvariables: []\nsync: {}\n"; got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	var decoded Workspace
	read(t, got, &decoded)
	if order := decoded.ToEntity().EnvironmentOrder; order == nil || len(order) != 0 {
		t.Fatalf("EnvironmentOrder = %v, want an empty stored order", order)
	}
}

func TestWorkspaceEnvironmentOrderKeepsSiblingKeys(t *testing.T) {
	model := NewWorkspace(entity.NewWorkspace("01W", "Acme").WithEnvironmentOrder([]string{"01D"}))
	source := "xenon: 1\nkind: workspace\nid: 01W\nname: Acme\ndescription: \"\"\norder: []\nenvironments:\n  pinned: true\n  order:\n    - 01P\nvariables: []\nsync: {}\n"
	want := "xenon: 1\nkind: workspace\nid: 01W\nname: Acme\ndescription: \"\"\norder: []\nenvironments:\n  pinned: true\n  order:\n    - 01D\nvariables: []\nsync: {}\n"
	if got := patched(t, source, model.Fields()); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}
