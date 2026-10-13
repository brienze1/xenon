package entity

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/changetype"
	"github.com/brienze1/xenon/internal/domain/enums/conflictkind"
	"github.com/brienze1/xenon/internal/domain/enums/conflictresolution"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/domain/enums/linkmode"
	"github.com/brienze1/xenon/internal/domain/enums/syncstate"
	"github.com/brienze1/xenon/pkg/errs"
)

func TestSyncSettingsCleanTheSubfolderAndRejectEscapes(t *testing.T) {
	settings := NewSyncSettings(" https://git.test/a.git ", " main ", "\\api\\xenon/")
	if settings != (SyncSettings{Repository: "https://git.test/a.git", Branch: "main", Subfolder: "api/xenon"}) {
		t.Fatalf("settings = %+v", settings)
	}
	if NewSyncSettings("r", "b", "").Subfolder != "" || NewSyncSettings("r", "b", "./").Subfolder != "" {
		t.Fatal("the repository root is an empty subfolder")
	}
	if !(SyncSettings{}).IsEmpty() || settings.IsEmpty() || (SyncSettings{Subfolder: "x"}).IsEmpty() || (SyncSettings{Branch: "x"}).IsEmpty() {
		t.Fatal("IsEmpty")
	}
	if settings.Differs(settings) || !settings.Differs(SyncSettings{Repository: settings.Repository, Branch: "dev", Subfolder: settings.Subfolder}) {
		t.Fatal("Differs")
	}
	for subfolder, safe := range map[string]bool{"": true, "api/xenon": true, "a/..b/c": true, "../up": false, "a/../b": false, "/abs": false, "C:/x": false, "a\\..\\b": false} {
		if HasSafeSubfolder(subfolder) != safe {
			t.Errorf("HasSafeSubfolder(%q) = %v", subfolder, !safe)
		}
	}
}

func TestLocalSyncSettingsKeepTheLink(t *testing.T) {
	shared := NewSyncSettings("repo", "main", "api/xenon")
	settings := NewLocalSyncSettings(shared, "acc", filepath.Join("clones", "acme", ""))
	if !settings.IsLinked() || !settings.HasAccount() || settings.Shared() != shared {
		t.Fatalf("settings = %+v", settings)
	}
	if settings.WorkspacePath() != filepath.Join("clones", "acme", "api", "xenon") {
		t.Fatalf("workspace path = %s", settings.WorkspacePath())
	}
	updated := settings.WithOptions("other", true, true)
	if updated.AccountID != "other" || !updated.AutoPull || !updated.AutoPush || updated.Repository != "repo" {
		t.Fatalf("updated = %+v", updated)
	}
	if settings.WithoutAccount().HasAccount() {
		t.Fatal("WithoutAccount")
	}
	for _, unlinked := range []LocalSyncSettings{{}, {ClonePath: "c", Branch: "b"}, {ClonePath: "c", Repository: "r"}, {Branch: "b", Repository: "r"}} {
		if unlinked.IsLinked() {
			t.Errorf("%+v should not be linked", unlinked)
		}
	}
}

func TestSyncStatusStates(t *testing.T) {
	settings := LocalSyncSettings{Repository: "r", Branch: "main"}
	cases := []struct {
		heads   BranchHeads
		changes int
		want    string
	}{
		{BranchHeads{}, 0, string(syncstate.UpToDate)},
		{BranchHeads{}, 2, string(syncstate.LocalChanges)},
		{BranchHeads{Ahead: 1}, 2, string(syncstate.Ahead)},
		{BranchHeads{Behind: 1}, 2, string(syncstate.Behind)},
		{BranchHeads{Ahead: 1, Behind: 3}, 0, string(syncstate.Diverged)},
	}
	for _, c := range cases {
		status := NewSyncStatus(settings, c.heads, c.changes)
		if string(status.State) != c.want || status.LocalChanges != c.changes || status.Ahead != c.heads.Ahead || status.Behind != c.heads.Behind || status.Settings != settings {
			t.Errorf("status %+v, want state %s", status, c.want)
		}
	}
	at := time.Date(2026, 10, 11, 9, 0, 0, 0, time.UTC)
	synced := NewSyncStatus(settings, BranchHeads{}, 0).SyncedAt(at)
	if synced.LastSyncAt == nil || !synced.LastSyncAt.Equal(at) {
		t.Fatal("SyncedAt")
	}
	failed := synced.WithError(SyncError{Code: "GIT-10503"})
	if failed.State != syncstate.Error || failed.LastError.Code != "GIT-10503" {
		t.Fatalf("failed = %+v", failed)
	}
	if synced.InConflict().State != syncstate.Conflict || NotLinkedSyncStatus().State != syncstate.NotLinked || SyncingSyncStatus(settings).State != syncstate.Syncing {
		t.Fatal("special states")
	}
	if SyncingSyncStatus(settings).Settings != settings {
		t.Fatal("syncing keeps the settings")
	}
}

func TestSyncStatusNoticesChangedSharedSettings(t *testing.T) {
	settings := LocalSyncSettings{Repository: "r", Branch: "main", Subfolder: "api"}
	status := NewSyncStatus(settings, BranchHeads{}, 0)
	if status.WithSharedSettings(settings.Shared()).Notice != "" || status.WithSharedSettings(SyncSettings{}).Notice != "" {
		t.Fatal("no notice for the same or no shared settings")
	}
	notice := status.WithSharedSettings(SyncSettings{Repository: "r2", Branch: "dev", Subfolder: "services"}).Notice
	for _, part := range []string{"r2", "dev", "\"services\"", "re-link"} {
		if !strings.Contains(notice, part) {
			t.Errorf("notice %q misses %q", notice, part)
		}
	}
	if status.LogValue().(map[string]any)["state"] != syncstate.UpToDate {
		t.Fatal("LogValue")
	}
}

func TestSyncErrorTakesTheNextStep(t *testing.T) {
	withStep := errs.BadRequestError("offline", "GIT-10503", errs.ErrorDetails{Attribute: "next_step", Messages: []string{"Try again"}})
	if failure := NewSyncError(withStep); failure != (SyncError{Code: "GIT-10503", Message: "offline", NextStep: "Try again"}) {
		t.Fatalf("failure = %+v", failure)
	}
	other := errs.BadRequestError("bad", "GIT-1", errs.ErrorDetails{Attribute: "field", Messages: []string{"x"}})
	if failure := NewSyncError(other); failure.NextStep != "" || failure.Code != "GIT-1" {
		t.Fatalf("failure = %+v", failure)
	}
	empty := errs.BadRequestError("bad", "GIT-2", errs.ErrorDetails{Attribute: "next_step"})
	if NewSyncError(empty).NextStep != "" {
		t.Fatal("a next step without messages")
	}
	if failure := NewSyncError(errors.New("plain")); failure != (SyncError{Message: "plain"}) {
		t.Fatalf("plain failure = %+v", failure)
	}
}

func TestSyncedFilesIgnoreLocalOnlyFilesAndListChanges(t *testing.T) {
	worktree := NewSyncedFiles(map[string]string{
		"workspace.yaml": "w2", "a.yaml": "a", "new.yaml": "n",
		".xenon/state.json": "x", "c/.xenon/cache": "x", ".xenon-import-1/a.yaml": "x", "a.yaml.xenon-tmp": "x",
	})
	if len(worktree) != 3 {
		t.Fatalf("worktree = %v", worktree)
	}
	head := map[string]string{"workspace.yaml": "w1", "a.yaml": "a", "gone.yaml": "g"}
	want := []PathChange{
		{Path: "gone.yaml", Change: changetype.Deleted},
		{Path: "new.yaml", Change: changetype.Added},
		{Path: "workspace.yaml", Change: changetype.Modified},
	}
	if changes := worktree.ChangesFrom(head); !reflect.DeepEqual(changes, want) {
		t.Fatalf("changes = %+v", changes)
	}
	if !want[0].IsDeletion() || want[1].IsDeletion() {
		t.Fatal("IsDeletion")
	}
	if worktree.Differs(head, "a.yaml") || !worktree.Differs(head, "gone.yaml") {
		t.Fatal("Differs")
	}
}

func TestWorktreeUpdateKeepsUncommittedWorkAndReportsOverlaps(t *testing.T) {
	current := map[string]string{"same": "1", "theirs": "1", "removed": "1", "edited": "1", "both": "1", "deleted-mine": "1", "agree": "1"}
	target := map[string]string{"same": "1", "theirs": "2", "edited": "1", "both": "2", "deleted-mine": "2", "added": "9", "agree": "5"}
	worktree := SyncedFiles{"same": "1", "theirs": "1", "removed": "1", "edited": "7", "both": "3", "agree": "5", "local-new": "4"}
	update := NewWorktreeUpdate(current, target, worktree)
	if !reflect.DeepEqual(update.Writes, map[string]string{"theirs": "2", "added": "9"}) {
		t.Errorf("writes = %v", update.Writes)
	}
	if !reflect.DeepEqual(update.Removals, []string{"removed"}) {
		t.Errorf("removals = %v", update.Removals)
	}
	if !reflect.DeepEqual(update.Conflicts, []string{"both", "deleted-mine"}) {
		t.Errorf("conflicts = %v", update.Conflicts)
	}
	if !update.HasConflicts() || update.IsEmpty() {
		t.Fatal("predicates")
	}
	plans := update.ConflictPlans(current, target, worktree)
	if !reflect.DeepEqual(plans, []PathPlan{{Path: "both", Base: "1", Mine: "3", Theirs: "2"}, {Path: "deleted-mine", Base: "1", Theirs: "2"}}) {
		t.Errorf("plans = %+v", plans)
	}
	quiet := NewWorktreeUpdate(current, current, SyncedFiles(current))
	if !quiet.IsEmpty() || quiet.HasConflicts() {
		t.Fatalf("quiet = %+v", quiet)
	}
	if NewWorktreeUpdate(map[string]string{}, map[string]string{"x": "1"}, SyncedFiles{}).IsEmpty() {
		t.Fatal("a write is not empty")
	}
	if NewWorktreeUpdate(map[string]string{"x": "1"}, map[string]string{}, SyncedFiles{"x": "1"}).IsEmpty() {
		t.Fatal("a removal is not empty")
	}
}

func TestChangeSetSelectsItemsAndBuildsTheTree(t *testing.T) {
	set := NewChangeSet([]ChangeItem{
		{ItemID: "r1", Change: changetype.Modified, Paths: []string{"a.yaml"}},
		{ItemID: "r2", Change: changetype.Deleted, Paths: []string{"b.yaml"}},
		{ItemID: "r3", Change: changetype.Renamed, Paths: []string{"old.yaml", "new.yaml"}},
		{ItemID: "r4", Change: changetype.Added, Paths: []string{"skip.yaml"}},
	})
	if set.Count() != 4 || set.IsEmpty() || !set.Selected(nil).IsEmpty() {
		t.Fatal("counts")
	}
	selected := set.Selected([]string{"r1", "r2", "r3", "unknown"})
	if selected.Count() != 3 || !reflect.DeepEqual(selected.Paths(), []string{"a.yaml", "b.yaml", "old.yaml", "new.yaml"}) {
		t.Fatalf("selected = %+v", selected)
	}
	head := map[string]string{"a.yaml": "a1", "b.yaml": "b1", "old.yaml": "o1", "keep.yaml": "k1"}
	worktree := SyncedFiles{"a.yaml": "a2", "new.yaml": "o1", "keep.yaml": "k1", "skip.yaml": "s1"}
	if tree := selected.TreeFrom(head, worktree); !reflect.DeepEqual(tree, map[string]string{"a.yaml": "a2", "new.yaml": "o1", "keep.yaml": "k1"}) {
		t.Fatalf("tree = %v", tree)
	}
	if head["a.yaml"] != "a1" {
		t.Fatal("the head must not change")
	}
	if present := selected.PresentPaths(worktree); !reflect.DeepEqual(present, []string{"a.yaml", "new.yaml"}) {
		t.Fatalf("present = %v", present)
	}
}

func TestChangeItemKeepsTheChangeOfSameKindPaths(t *testing.T) {
	added := NewChangeItem(ItemRef{ID: "w"}, PathChange{Path: "workspace.yaml", Change: changetype.Added}).Combined(PathChange{Path: ".gitignore", Change: changetype.Added})
	if added.Change != changetype.Added || len(added.Paths) != 2 {
		t.Fatalf("added = %+v", added)
	}
	if !added.IsAnyOf([]string{"x", "w"}) || added.IsAnyOf([]string{"x"}) {
		t.Fatal("IsAnyOf")
	}
}

func TestThreeWayPlanMergedTakesStructuralResults(t *testing.T) {
	plan := ThreeWayPlan{Paths: []PathPlan{
		NewPathPlan("a.yaml", "1", "1", "2"),
		NewPathPlan("b.yaml", "1", "3", "1"),
		NewPathPlan("gone.yaml", "1", "1", ""),
		NewPathPlan("collection.yaml", "1", "2", "3"),
	}}
	merged := plan.Merged(map[string]string{"collection.yaml": "m"})
	if !reflect.DeepEqual(merged, map[string]string{"a.yaml": "2", "b.yaml": "3", "collection.yaml": "m"}) {
		t.Fatalf("merged = %v", merged)
	}
}

func TestPathPlanAnyVersionPrefersTheirs(t *testing.T) {
	for plan, want := range map[PathPlan]string{
		{Theirs: "t", Base: "b", Mine: "m"}: "t",
		{Base: "b", Mine: "m"}:              "b",
		{Mine: "m"}:                         "m",
		{}:                                  "",
	} {
		if plan.AnyVersion() != want {
			t.Errorf("AnyVersion(%+v) = %q", plan, plan.AnyVersion())
		}
	}
}

func TestConflictItemsAndSessions(t *testing.T) {
	ref := ItemRef{ID: "r1", Name: "Users", Kind: itemkind.Request}
	both := NewConflictItem(ref, PathPlan{Path: "u.yaml", Base: "1", Mine: "2", Theirs: "3"}, nil)
	if both.Kind != conflictkind.BothModified || both.Resolution != conflictresolution.Unresolved || both.ItemName != "Users" || both.Path != "u.yaml" || both.Mine != "2" {
		t.Fatalf("both = %+v", both)
	}
	if NewConflictItem(ref, PathPlan{Mine: "2"}, nil).Kind != conflictkind.DeleteModify || NewConflictItem(ref, PathPlan{Theirs: "2"}, nil).Kind != conflictkind.DeleteModify {
		t.Fatal("delete vs modify")
	}
	if NewConflictItem(ref, PathPlan{Mine: "1", Theirs: "2"}, []FieldConflict{{Path: "name"}}).Kind != conflictkind.Field {
		t.Fatal("field conflicts")
	}
	at := time.Date(2026, 10, 11, 9, 0, 0, 0, time.UTC)
	other := NewConflictItem(ItemRef{ID: "r2"}, PathPlan{Mine: "1", Theirs: "2"}, nil)
	session := NewSyncSession("s1", "w1", BranchHeads{Local: "l", Remote: "r", Base: "b"}, true, []ConflictItem{both, other, both}, at)
	if session.ID != "s1" || session.LocalHead != "l" || session.RemoteHead != "r" || session.Base != "b" || !session.StartedAt.Equal(at) || len(session.Conflicts) != 2 {
		t.Fatalf("session = %+v", session)
	}
}

func TestGitAttributesDetectLargeFileStorage(t *testing.T) {
	for path, relevant := range map[string]bool{
		".gitattributes": true, "api/.gitattributes": true, "api/xenon/.gitattributes": true, "api/xenon/deep/.gitattributes": true,
		"web/.gitattributes": false, "apix/.gitattributes": false, "api/xenon/file.yaml": false,
	} {
		if IsGitAttributesFor(path, "api/xenon") != relevant {
			t.Errorf("IsGitAttributesFor(%q) = %v", path, !relevant)
		}
	}
	if !IsGitAttributesFor("web/.gitattributes", "") {
		t.Fatal("the repository root sees every attributes file")
	}
	if !(GitAttributes{"a": "*.txt text\napi/** filter=lfs diff=lfs"}).UsesLfs() {
		t.Fatal("lfs filter")
	}
	if (GitAttributes{"a": "# *.bin filter=lfs\n*.txt text"}).UsesLfs() || (GitAttributes{}).UsesLfs() {
		t.Fatal("comments and plain attributes are not lfs")
	}
}

func TestLinkModeAndRegistryOverlaps(t *testing.T) {
	workspace := FolderInfo{Exists: true, IsDirectory: true, HasWorkspaceFile: true}
	if workspace.LinkModeFor(linkmode.AdoptRepo) != linkmode.AdoptRepo || workspace.LinkModeFor(linkmode.Unspecified) != linkmode.Unspecified {
		t.Fatal("an existing workspace keeps the requested mode")
	}
	if (FolderInfo{}).LinkModeFor(linkmode.AdoptRepo) != linkmode.PlaceMine {
		t.Fatal("an empty subfolder places mine")
	}
	registry := NewAppRegistry()
	registry.Register(NewRegisteredWorkspace("w1", filepath.Join("root", "acme"), "Acme", time.Time{}))
	if !registry.OverlapsAny(filepath.Join("root", "acme", "clone")) || !registry.OverlapsAny("root") || registry.OverlapsAny(filepath.Join("root", "clones")) {
		t.Fatal("OverlapsAny")
	}
}

func TestLinkedWorkspaceAndSyncCommit(t *testing.T) {
	settings := LocalSyncSettings{ClonePath: "clone", Branch: "main", Subfolder: "api", Repository: "r"}
	linked := NewLinkedWorkspace("w1", "root", settings).WithAccess(GitAccount{CommitName: "Luis", CommitEmail: "l@a.test"}, GitAccess{Repository: "r"})
	if linked.ClonePath() != "clone" || linked.Branch() != "main" || linked.Subfolder() != "api" || linked.Access.Repository != "r" {
		t.Fatalf("linked = %+v", linked)
	}
	if linked.LogValue().(map[string]any)["root"] != "root" {
		t.Fatal("LogValue")
	}
	at := time.Date(2026, 10, 11, 9, 0, 0, 0, time.UTC)
	commit := NewSyncCommit(linked, "parent", "message", map[string]string{"a": "1"}, at)
	want := SyncCommit{When: at, Blobs: map[string]string{"a": "1"}, Branch: "main", Subfolder: "api", Parent: "parent", Message: "message", AuthorName: "Luis", AuthorEmail: "l@a.test"}
	if !reflect.DeepEqual(commit, want) {
		t.Fatalf("commit = %+v", commit)
	}
	if commit.IsEmpty() || (SyncCommit{Files: map[string]string{"a": "x"}}).IsEmpty() || !(SyncCommit{}).IsEmpty() {
		t.Fatal("IsEmpty")
	}
	if commit.LogValue().(map[string]any)["blobs"] != 1 {
		t.Fatal("LogValue")
	}
	if !(BranchHeads{Local: "x"}).HasLocalCommits() || (BranchHeads{}).HasLocalCommits() {
		t.Fatal("HasLocalCommits")
	}
}
