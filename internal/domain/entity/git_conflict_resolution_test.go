package entity

import (
	"reflect"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/conflictkind"
	"github.com/brienze1/xenon/internal/domain/enums/conflictresolution"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func TestConflictItemAcceptsOnlyTheChoicesOfItsKind(t *testing.T) {
	request := NewConflictItem(ItemRef{ID: "r", Name: "Users", Kind: itemkind.Request}, PathPlan{Path: "u.yaml", Base: "1", Mine: "2", Theirs: "3"}, nil)
	environment := NewConflictItem(ItemRef{ID: "e", Kind: itemkind.Environment}, PathPlan{Base: "1", Mine: "2", Theirs: "3"}, nil)
	collection := NewConflictItem(ItemRef{ID: "c", Kind: itemkind.Collection}, PathPlan{Base: "1", Mine: "2", Theirs: "3"}, nil)
	deletion := NewConflictItem(ItemRef{ID: "d", Kind: itemkind.Request}, PathPlan{Base: "1", Mine: "2"}, nil)
	fields := NewConflictItem(ItemRef{ID: "f", Kind: itemkind.Folder}, PathPlan{Base: "1", Mine: "2", Theirs: "3"}, []FieldConflict{{Path: "name"}, {Path: "description"}})
	cases := []struct {
		item       ConflictItem
		resolution enums.ConflictResolution
		field      string
		accepted   bool
	}{
		{request, conflictresolution.Mine, "", true},
		{request, conflictresolution.Theirs, "", true},
		{request, conflictresolution.Both, "", true},
		{request, conflictresolution.Keep, "", false},
		{request, conflictresolution.Delete, "", false},
		{request, conflictresolution.Mine, "name", false},
		{environment, conflictresolution.Both, "", true},
		{collection, conflictresolution.Both, "", false},
		{collection, conflictresolution.Theirs, "", true},
		{deletion, conflictresolution.Keep, "", true},
		{deletion, conflictresolution.Delete, "", true},
		{deletion, conflictresolution.Mine, "", false},
		{deletion, conflictresolution.Both, "", false},
		{deletion, conflictresolution.Keep, "name", false},
		{fields, conflictresolution.Mine, "name", true},
		{fields, conflictresolution.Theirs, "description", true},
		{fields, conflictresolution.Mine, "", false},
		{fields, conflictresolution.Mine, "other", false},
		{fields, conflictresolution.Both, "name", false},
		{fields, conflictresolution.Keep, "name", false},
	}
	for _, c := range cases {
		if c.item.Accepts(c.resolution, c.field) != c.accepted {
			t.Errorf("%s/%s accepts %s %q = %v", c.item.ItemID, c.item.Kind, c.resolution, c.field, !c.accepted)
		}
	}
}

func TestConflictItemResultsAndCopies(t *testing.T) {
	request := NewConflictItem(ItemRef{ID: "r", Name: "Users", Kind: itemkind.Request}, PathPlan{Path: "u.yaml", Base: "1", Mine: "2", Theirs: "3"}, nil)
	if request.IsResolved() || request.KeepsBoth() || request.IsFieldMerge() || !request.CanKeepBoth() {
		t.Fatalf("fresh conflict %+v", request)
	}
	for resolution, want := range map[enums.ConflictResolution]string{conflictresolution.Mine: "2", conflictresolution.Theirs: "3", conflictresolution.Both: "2"} {
		resolved := request.Resolved(resolution, "")
		if !resolved.IsResolved() || resolved.Result() != want || resolved.KeepsBoth() != (resolution == conflictresolution.Both) {
			t.Errorf("%s: result %q, resolved %+v", resolution, resolved.Result(), resolved)
		}
	}
	mineDeleted := NewConflictItem(ItemRef{ID: "d"}, PathPlan{Base: "1", Theirs: "3"}, nil)
	theirsDeleted := NewConflictItem(ItemRef{ID: "d"}, PathPlan{Base: "1", Mine: "2"}, nil)
	if mineDeleted.Resolved(conflictresolution.Keep, "").Result() != "3" || theirsDeleted.Resolved(conflictresolution.Keep, "").Result() != "2" {
		t.Fatal("keep takes the modified side")
	}
	if theirsDeleted.Resolved(conflictresolution.Delete, "").Result() != "" {
		t.Fatal("delete removes the file")
	}
	if request.CopyName() != "Users (theirs)" {
		t.Fatalf("copy name %q", request.CopyName())
	}
	uncommitted := NewUncommittedConflictItem(ItemRef{ID: "u", Kind: itemkind.Folder}, PathPlan{Base: "1", Mine: "2", Theirs: "3"})
	if !uncommitted.Uncommitted || uncommitted.Kind != conflictkind.BothModified || uncommitted.CanKeepBoth() || request.Uncommitted {
		t.Fatalf("uncommitted %+v", uncommitted)
	}
	versions := ConflictVersions{Mine: &ConflictVersion{Text: "m"}}
	if request.WithVersions(versions).Versions.Mine.Text != "m" || request.Versions.Mine != nil {
		t.Fatal("WithVersions copies")
	}
}

func TestFieldConflictsResolveOneFieldAtATime(t *testing.T) {
	mine, theirs := "Mine", "Theirs"
	field := FieldConflict{Path: "name", Resolution: conflictresolution.Unresolved}.WithTexts(&mine, &theirs)
	if *field.MineText != "Mine" || *field.TheirsText != "Theirs" {
		t.Fatalf("texts %+v", field)
	}
	item := NewConflictItem(ItemRef{ID: "f"}, PathPlan{Mine: "1", Theirs: "2"}, []FieldConflict{field, {Path: "description", Resolution: conflictresolution.Unresolved}})
	if !item.IsFieldMerge() {
		t.Fatal("field merge")
	}
	partly := item.Resolved(conflictresolution.Theirs, "name")
	if partly.IsResolved() || partly.Fields[0].Resolution != conflictresolution.Theirs || item.Fields[0].Resolution != conflictresolution.Unresolved {
		t.Fatalf("partly %+v", partly)
	}
	done := partly.Resolved(conflictresolution.Mine, "description")
	if !done.IsResolved() || done.Resolution != conflictresolution.Merged || done.Result() != "1" {
		t.Fatalf("done %+v", done)
	}
	want := map[string]enums.ConflictResolution{"name": conflictresolution.Theirs, "description": conflictresolution.Mine}
	if !reflect.DeepEqual(done.FieldResolutions(), want) {
		t.Fatalf("field resolutions %v", done.FieldResolutions())
	}
	if len(item.WithFields(nil).Fields) != 0 || len(item.Fields) != 2 {
		t.Fatal("WithFields copies")
	}
}

func TestSyncSessionTracksResolutions(t *testing.T) {
	users := NewConflictItem(ItemRef{ID: "users", Kind: itemkind.Request}, PathPlan{Path: "c/users.yaml", Base: "1", Mine: "2", Theirs: "3"}, nil)
	gone := NewConflictItem(ItemRef{ID: "gone", Kind: itemkind.Request}, PathPlan{Path: "c/gone.yaml", Base: "1", Mine: "2"}, nil)
	folder := NewConflictItem(ItemRef{ID: "folder", Kind: itemkind.Folder}, PathPlan{Path: "c/f/folder.yaml", Base: "1", Mine: "2", Theirs: "3"}, []FieldConflict{{Path: "name", Resolution: conflictresolution.Unresolved}})
	draft := NewUncommittedConflictItem(ItemRef{ID: "draft"}, PathPlan{Path: "c/draft.yaml", Base: "1", Mine: "2", Theirs: "3"})
	heads := BranchHeads{Local: "l", Remote: "r", Base: "b", Ahead: 1}
	session := NewSyncSession("s", "w", heads, true, []ConflictItem{users, gone, folder, draft}, time.Now())
	if session.LocalID != "w" || !session.Replay || session.Heads() != (BranchHeads{Local: "l", Remote: "r", Base: "b"}) || !session.StartedFrom("l") || session.StartedFrom("x") {
		t.Fatalf("session %+v", session)
	}
	if _, found := session.Conflict("missing"); found {
		t.Fatal("missing conflict")
	}
	if conflict, found := session.Conflict("gone"); !found || conflict.Path != "c/gone.yaml" {
		t.Fatalf("conflict %+v", conflict)
	}
	if len(session.CommittedConflicts()) != 3 || len(session.UncommittedConflicts()) != 1 || session.UncommittedConflicts()[0].ItemID != "draft" {
		t.Fatal("committed and uncommitted conflicts")
	}
	resolved := session.WithConflict(users.Resolved(conflictresolution.Theirs, "")).
		WithConflict(gone.Resolved(conflictresolution.Delete, "")).
		WithConflict(folder.Resolved(conflictresolution.Mine, "name"))
	if resolved.IsResolved() || session.Conflicts[0].IsResolved() {
		t.Fatal("draft is still unresolved and the original session is unchanged")
	}
	if !resolved.WithConflict(draft.Resolved(conflictresolution.Mine, "")).IsResolved() {
		t.Fatal("every conflict resolved")
	}
	if !reflect.DeepEqual(resolved.FieldResolutions(), map[string]map[string]enums.ConflictResolution{"c/f/folder.yaml": {"name": conflictresolution.Mine}}) {
		t.Fatalf("field resolutions %v", resolved.FieldResolutions())
	}
	tree := map[string]string{"c/users.yaml": "2", "c/gone.yaml": "2", "c/f/folder.yaml": "merged", "c/draft.yaml": "4", "c/other.yaml": "5"}
	want := map[string]string{"c/users.yaml": "3", "c/f/folder.yaml": "merged", "c/draft.yaml": "4", "c/other.yaml": "5"}
	if got := resolved.ResolvedTree(tree); !reflect.DeepEqual(got, want) || tree["c/users.yaml"] != "2" {
		t.Fatalf("resolved tree %v", got)
	}
	if resolved.LogValue().(map[string]any)["conflicts"] != 4 {
		t.Fatal("log value")
	}
}

func TestWorktreeUpdateTakesTheResolvedContent(t *testing.T) {
	worktree := SyncedFiles{"a.yaml": "mine", "b.yaml": "same"}
	update := WorktreeUpdate{Writes: map[string]string{"a.yaml": "theirs", "c.yaml": "new"}, Removals: []string{"a.yaml", "d.yaml"}, Conflicts: []string{"a.yaml", "e.yaml"}}
	kept := update.Resolved("a.yaml", "mine", worktree)
	if _, written := kept.Writes["a.yaml"]; written || kept.Writes["c.yaml"] != "new" || !reflect.DeepEqual(kept.Removals, []string{"d.yaml"}) || !reflect.DeepEqual(kept.Conflicts, []string{"e.yaml"}) {
		t.Fatalf("kept %+v", kept)
	}
	if theirs := update.Resolved("a.yaml", "theirs", worktree); theirs.Writes["a.yaml"] != "theirs" {
		t.Fatalf("theirs %+v", theirs)
	}
	if removed := update.Resolved("a.yaml", "", worktree); !reflect.DeepEqual(removed.Removals, []string{"d.yaml", "a.yaml"}) {
		t.Fatalf("removed %+v", removed)
	}
	if absent := update.Resolved("z.yaml", "", worktree); len(absent.Removals) != 2 || len(absent.Writes) != 2 {
		t.Fatalf("an absent file stays absent %+v", absent)
	}
	if update.Writes["a.yaml"] != "theirs" || len(update.Removals) != 2 {
		t.Fatal("the original update is unchanged")
	}
}

func TestDocumentIdentityAndOrderEntries(t *testing.T) {
	document := NewMappingNode().Set("kind", NewScalarNode("!!str", "request")).Set("id", NewScalarNode("!!str", "old")).Set("name", NewScalarNode("!!str", "Users"))
	renamed := document.Renamed("new", "Users (theirs)")
	if renamed.Field("id").Value != "new" || renamed.Field("name").Value != "Users (theirs)" || document.Field("id").Value != "old" {
		t.Fatalf("renamed %+v", renamed)
	}
	if !reflect.DeepEqual(renamed.Keys, []string{"kind", "id", "name"}) {
		t.Fatalf("keys %v", renamed.Keys)
	}
	order := NewSequenceNode([]*DocumentNode{NewScalarNode("!!str", "a.yaml"), NewScalarNode("!!str", "b.yaml")})
	parent := NewMappingNode().Set("name", NewScalarNode("!!str", "API")).Set("order", order)
	listed, changed := parent.WithOrderEntryAfter("a.yaml", "a-copy.yaml")
	if values, _ := listed.Field("order").ScalarItems(); !changed || !reflect.DeepEqual(values, []string{"a.yaml", "a-copy.yaml", "b.yaml"}) {
		t.Fatalf("order %v", values)
	}
	if values, _ := parent.Field("order").ScalarItems(); len(values) != 2 || listed.Field("name").Value != "API" {
		t.Fatal("the original document is unchanged")
	}
	if last, changed := parent.WithOrderEntryAfter("b.yaml", "b-copy.yaml"); !changed || last.Field("order").Items[2].Value != "b-copy.yaml" {
		t.Fatal("insert after the last entry")
	}
	if same, changed := parent.WithOrderEntryAfter("missing.yaml", "x.yaml"); changed || same != parent {
		t.Fatal("an unlisted file leaves the order alone")
	}
	if same, changed := NewMappingNode().WithOrderEntryAfter("a.yaml", "x.yaml"); changed || same == nil {
		t.Fatal("no order list")
	}
	if !NewScalarNode("!!str", "x").IsScalar() || parent.IsScalar() || (*DocumentNode)(nil).IsScalar() {
		t.Fatal("IsScalar")
	}
}
