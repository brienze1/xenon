package usecase

import (
	"reflect"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/conflictresolution"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/domain/enums/pathaction"
)

func docScalar(value string) *entity.DocumentNode {
	return entity.NewScalarNode("!!str", value)
}

func docList(values ...string) *entity.DocumentNode {
	items := make([]*entity.DocumentNode, 0, len(values))
	for _, value := range values {
		items = append(items, docScalar(value))
	}
	return entity.NewSequenceNode(items)
}

func doc(pairs ...any) *entity.DocumentNode {
	node := entity.NewMappingNode()
	for index := 0; index < len(pairs); index += 2 {
		value := pairs[index+1]
		switch typed := value.(type) {
		case string:
			node.Set(pairs[index].(string), docScalar(typed))
		case *entity.DocumentNode:
			node.Set(pairs[index].(string), typed)
		}
	}
	return node
}

func docRows(entries ...*entity.DocumentNode) *entity.DocumentNode {
	return entity.NewSequenceNode(entries)
}

func docRow(key, value string) *entity.DocumentNode {
	return doc("key", key, "value", value)
}

func TestPlanThreeWayClassifiesEveryPathInOrder(t *testing.T) {
	base := map[string]string{"a.yaml": "1", "b.yaml": "1", "c.yaml": "1", "d.yaml": "1", "e.yaml": "1", "collections/x/collection.yaml": "1", "g.yaml": "1"}
	mine := map[string]string{"a.yaml": "1", "b.yaml": "2", "c.yaml": "2", "d.yaml": "2", "collections/x/collection.yaml": "2", "f.yaml": "9", "g.yaml": "3"}
	theirs := map[string]string{"a.yaml": "1", "b.yaml": "1", "c.yaml": "3", "d.yaml": "2", "e.yaml": "4", "collections/x/collection.yaml": "3", "g.yaml": "1", "h.yaml": "5"}
	plan := PlanThreeWay(base, mine, theirs)
	got := map[string]enums.PathAction{}
	order := make([]string, 0)
	for _, path := range plan.Paths {
		got[path.Path] = path.Action
		order = append(order, path.Path)
	}
	want := map[string]enums.PathAction{
		"a.yaml": pathaction.Keep, "b.yaml": pathaction.TakeMine, "c.yaml": pathaction.BothModified,
		"d.yaml": pathaction.Keep, "e.yaml": pathaction.DeleteModify, "collections/x/collection.yaml": pathaction.MergeStructural,
		"f.yaml": pathaction.TakeMine, "g.yaml": pathaction.TakeMine, "h.yaml": pathaction.TakeTheirs,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("actions = %v, want %v", got, want)
	}
	if !reflect.DeepEqual(order, []string{"a.yaml", "b.yaml", "c.yaml", "collections/x/collection.yaml", "d.yaml", "e.yaml", "f.yaml", "g.yaml", "h.yaml"}) {
		t.Fatalf("paths are not sorted: %v", order)
	}
	if conflicts := plan.Conflicts(); len(conflicts) != 2 || conflicts[0].Path != "c.yaml" || conflicts[1].Path != "e.yaml" {
		t.Fatalf("Conflicts() = %+v", conflicts)
	}
	if merges := plan.StructuralMerges(); len(merges) != 1 || merges[0].Path != "collections/x/collection.yaml" {
		t.Fatalf("StructuralMerges() = %+v", merges)
	}
}

func TestSquashReplayMessage(t *testing.T) {
	cases := map[string]struct {
		messages []string
		want     string
	}{
		"nothing":  {nil, "Sync"},
		"blank":    {[]string{" ", "\n"}, "Sync"},
		"one":      {[]string{"Add orders\n\nwith body"}, "Sync: Add orders\n\n- Add orders\n  \n  with body"},
		"several":  {[]string{"Add orders", "", "Fix auth", "Rename ping"}, "Sync: Add orders (+2 more)\n\n- Add orders\n- Fix auth\n- Rename ping"},
		"two only": {[]string{"  Add orders  ", "Fix auth"}, "Sync: Add orders (+1 more)\n\n- Add orders\n- Fix auth"},
	}
	for name, c := range cases {
		if got := SquashReplayMessage(c.messages); got != c.want {
			t.Errorf("%s: SquashReplayMessage = %q, want %q", name, got, c.want)
		}
	}
}

func TestMergeStructuralFileMergesNonOverlappingFields(t *testing.T) {
	base := doc("name", "Billing", "description", "old", "order", docList("a.yaml", "b.yaml", "c.yaml"), "auth", doc("type", "bearer", "bearer", doc("token", "{{t}}")))
	mine := doc("name", "Billing v2", "description", "old", "order", docList("a.yaml", "c.yaml", "mine.yaml"), "auth", doc("type", "bearer", "bearer", doc("token", "{{t}}")), "docs", "mine")
	theirs := doc("name", "Billing", "description", "new", "order", docList("c.yaml", "b.yaml", "theirs.yaml", "a.yaml"), "auth", doc("type", "bearer", "bearer", doc("token", "{{t2}}")))
	merged := MergeStructuralFile(itemkind.Collection, base, mine, theirs, nil)
	if merged.HasConflicts() || len(merged.Notes) != 0 {
		t.Fatalf("unexpected conflicts %+v", merged)
	}
	want := doc("name", "Billing v2", "description", "new", "order", docList("c.yaml", "theirs.yaml", "a.yaml", "mine.yaml"), "auth", doc("type", "bearer", "bearer", doc("token", "{{t2}}")), "docs", "mine")
	if !merged.Result.Equals(want) {
		t.Fatalf("result %+v", merged.Result)
	}
	if !reflect.DeepEqual(merged.Result.Keys, []string{"name", "description", "order", "auth", "docs"}) {
		t.Fatalf("keys %v", merged.Result.Keys)
	}
}

func TestMergeStructuralFileReportsFieldConflictsAndKeepsMinePlaceholder(t *testing.T) {
	base := doc("name", "Billing", "auth", doc("bearer", doc("token", "a")), "custom", "x", "variables", docRows(docRow("v1", "1")))
	mine := doc("name", "Mine", "auth", doc("bearer", doc("token", "b")), "custom", "y", "variables", docRows(docRow("v1", "2")))
	theirs := doc("name", "Theirs", "auth", doc("bearer", doc("token", "c")), "custom", "z", "variables", docRows(docRow("v1", "3")))
	merged := MergeStructuralFile(itemkind.Folder, base, mine, theirs, nil)
	paths := make([]string, 0)
	for _, conflict := range merged.FieldConflicts {
		paths = append(paths, conflict.Path)
		if conflict.Resolution != conflictresolution.Unresolved {
			t.Errorf("%s resolution = %s", conflict.Path, conflict.Resolution)
		}
	}
	if !reflect.DeepEqual(paths, []string{"name", "auth.bearer.token", "variables[v1].value"}) {
		t.Fatalf("conflict paths = %v", paths)
	}
	if merged.FieldConflicts[0].Mine.Value != "Mine" || merged.FieldConflicts[0].Theirs.Value != "Theirs" {
		t.Fatalf("conflict versions %+v", merged.FieldConflicts[0])
	}
	if !reflect.DeepEqual(merged.Notes, []string{"kept their custom"}) || merged.Result.Field("custom").Value != "z" {
		t.Fatalf("unknown field handling: notes %v, custom %+v", merged.Notes, merged.Result.Field("custom"))
	}
	if merged.Result.Field("name").Value != "Mine" || merged.ItemConflict || !merged.HasConflicts() {
		t.Fatalf("placeholder %+v", merged)
	}
}

func TestMergeStructuralFileMergesKeyedRowsAndDeletions(t *testing.T) {
	base := doc("variables", docRows(docRow("keep", "1"), docRow("dropped-by-mine", "1"), docRow("dropped-by-theirs", "1"), docRow("changed", "1")), "description", "d")
	mine := doc("variables", docRows(docRow("keep", "1"), docRow("dropped-by-theirs", "1"), docRow("changed", "1"), docRow("added-by-mine", "m")))
	theirs := doc("variables", docRows(docRow("added-by-theirs", "t"), docRow("keep", "1"), docRow("dropped-by-mine", "1"), docRow("changed", "2")), "description", "d")
	merged := MergeStructuralFile(itemkind.Environment, base, mine, theirs, nil)
	if merged.HasConflicts() {
		t.Fatalf("unexpected conflicts %+v", merged)
	}
	want := doc("variables", docRows(docRow("added-by-theirs", "t"), docRow("keep", "1"), docRow("changed", "2"), docRow("added-by-mine", "m")))
	if !merged.Result.Equals(want) {
		keys, _, _ := merged.Result.Field("variables").KeyedRows()
		t.Fatalf("rows %v, description %+v", keys, merged.Result.Field("description"))
	}
}

func TestMergeStructuralFileMakesEnvironmentRowConflictsItemConflicts(t *testing.T) {
	base := doc("name", "hml", "variables", docRows(docRow("url", "a")))
	mine := doc("name", "hml", "variables", docRows(docRow("url", "b")))
	theirs := doc("name", "hml", "variables", docRows(docRow("url", "c")))
	merged := MergeStructuralFile(itemkind.Environment, base, mine, theirs, nil)
	if !merged.ItemConflict || len(merged.FieldConflicts) != 0 || !merged.HasConflicts() {
		t.Fatalf("expected an item conflict, got %+v", merged)
	}
}

func TestMergeStructuralFileFallsBackToConflictsForUnmergeableShapes(t *testing.T) {
	duplicated := docRows(docRow("h", "1"), docRow("h", "2"))
	cases := map[string]struct {
		base, mine, theirs *entity.DocumentNode
		path               string
	}{
		"duplicate row keys":         {doc("headers", docRows(docRow("h", "0"))), doc("headers", duplicated), doc("headers", docRows(docRow("h", "3"))), "headers"},
		"order with mappings":        {doc("order", docList("a")), doc("order", docRows(docRow("a", "1"))), doc("order", docList("b")), "order"},
		"order over a mapping base":  {doc("order", doc("x", "1")), doc("order", docList("a")), doc("order", docList("b")), "order"},
		"rows over a scalar base":    {doc("headers", "none"), doc("headers", docRows(docRow("a", "1"))), doc("headers", docRows(docRow("b", "1"))), "headers"},
		"mapping against scalar":     {doc("auth", "none"), doc("auth", doc("type", "basic")), doc("auth", "noauth"), "auth"},
		"mapping over scalar base":   {doc("auth", "none"), doc("auth", doc("type", "basic")), doc("auth", doc("type", "bearer")), "auth"},
		"deleted against changed":    {doc("description", "a"), doc(), doc("description", "b"), "description"},
		"rows against scalar values": {doc("variables", docRows(docRow("a", "1"))), doc("variables", docRows(docRow("a", "2"))), doc("variables", "none"), "variables"},
	}
	for name, c := range cases {
		merged := MergeStructuralFile(itemkind.Collection, c.base, c.mine, c.theirs, nil)
		if len(merged.FieldConflicts) != 1 || merged.FieldConflicts[0].Path != c.path {
			t.Errorf("%s: conflicts %+v", name, merged.FieldConflicts)
		}
	}
}

func TestMergeStructuralFileMergesAddedFilesWithoutBase(t *testing.T) {
	mine := doc("name", "N", "order", docList("a.yaml"), "variables", docRows(docRow("x", "1")))
	theirs := doc("name", "N", "order", docList("b.yaml"), "variables", docRows(docRow("y", "1")))
	merged := MergeStructuralFile(itemkind.Folder, nil, mine, theirs, nil)
	want := doc("name", "N", "order", docList("b.yaml", "a.yaml"), "variables", docRows(docRow("y", "1"), docRow("x", "1")))
	if merged.HasConflicts() || !merged.Result.Equals(want) {
		t.Fatalf("merge without base = %+v", merged)
	}
}

func TestMergeStructuralFileReportsRootConflicts(t *testing.T) {
	merged := MergeStructuralFile(itemkind.Workspace, docScalar("a"), docScalar("b"), docScalar("c"), nil)
	if len(merged.FieldConflicts) != 1 || merged.FieldConflicts[0].Path != "" || merged.Result.Value != "b" {
		t.Fatalf("root conflict %+v", merged)
	}
}

func TestOrderMergeDeduplicates(t *testing.T) {
	merged, ok := mergeOrder(docList("a"), docList("a", "x", "x", "y"), docList("y", "y", "a", "z"))
	values, _ := merged.ScalarItems()
	if !ok || strings.Join(values, ",") != "y,a,z,x" {
		t.Fatalf("mergeOrder = %v, %v", values, ok)
	}
}
