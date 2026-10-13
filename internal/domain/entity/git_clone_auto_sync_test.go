package entity

import (
	"path/filepath"
	"testing"
)

func TestSyncSettingsBranchAndWorkspaceName(t *testing.T) {
	settings := NewSyncSettings("https://git.test/acme/backend.git", "", "api")
	if settings.HasBranch() || !settings.OnBranch("trunk").HasBranch() {
		t.Fatal("HasBranch")
	}
	if onTrunk := settings.OnBranch("trunk"); onTrunk.Branch != "trunk" || onTrunk.Repository != settings.Repository || onTrunk.Subfolder != "api" || settings.Branch != "" {
		t.Fatalf("OnBranch = %+v", onTrunk)
	}
	for repository, name := range map[string]string{
		"https://git.test/acme/backend.git":     "backend",
		"https://git.test/acme/backend.git/":    "backend",
		"git@github.com:acme/tools.git":         "tools",
		"ssh://git@host:22/team/api":            "api",
		"C:\\repos\\local.git":                  "local",
		"https://dev.azure.com/org/proj/_git/x": "x",
		"bare":                                  "bare",
	} {
		if got := (SyncSettings{Repository: repository}).WorkspaceNameOr("  "); got != name {
			t.Errorf("WorkspaceNameOr(%q) = %q, want %q", repository, got, name)
		}
	}
	if got := settings.WorkspaceNameOr("  Backend API "); got != "Backend API" {
		t.Fatalf("a given name wins, got %q", got)
	}
}

func TestRemoteInfoCloneBranch(t *testing.T) {
	if (RemoteInfo{DefaultBranch: "trunk"}).CloneBranch() != "trunk" {
		t.Fatal("the remote's default branch is cloned")
	}
	if (RemoteInfo{}).CloneBranch() != "main" {
		t.Fatal("an empty remote falls back to main")
	}
}

func TestLinkedWorkspaceWithSettings(t *testing.T) {
	linked := NewLinkedWorkspace("L", "root", LocalSyncSettings{Branch: "main"})
	moved := linked.WithSettings(LocalSyncSettings{Branch: "trunk"})
	if moved.Branch() != "trunk" || linked.Branch() != "main" || moved.LocalID != "L" || moved.Root != "root" {
		t.Fatalf("moved = %+v, linked = %+v", moved, linked)
	}
}

func TestLocalSyncSettingsAutomaticSyncAndClone(t *testing.T) {
	linked := NewLocalSyncSettings(NewSyncSettings("repo", "main", "api"), "acc", filepath.Join("clones", "acme"))
	if linked.PullsWhenOpened() || linked.PushesAfterSave() {
		t.Fatal("both options are off by default")
	}
	both := linked.WithOptions("acc", true, true)
	if !both.PullsWhenOpened() || !both.PushesAfterSave() {
		t.Fatal("options on")
	}
	if linked.WithOptions("acc", true, false).PushesAfterSave() || linked.WithOptions("acc", false, true).PullsWhenOpened() {
		t.Fatal("each option is independent")
	}
	unlinked := LocalSyncSettings{AutoPull: true, AutoPush: true}
	if unlinked.PullsWhenOpened() || unlinked.PushesAfterSave() {
		t.Fatal("an unlinked workspace never syncs automatically")
	}
	for path, inside := range map[string]bool{
		filepath.Join("clones", "acme"):              true,
		filepath.Join("clones", "acme", "api", "x"):  true,
		filepath.Join("clones", "acme-plain"):        false,
		filepath.Join("clones"):                      false,
		filepath.Join("elsewhere", "clones", "acme"): false,
	} {
		if linked.Encloses(path) != inside {
			t.Errorf("Encloses(%q) = %v", path, !inside)
		}
	}
}

func TestChangeSetUpdateMessage(t *testing.T) {
	changes := NewChangeSet([]ChangeItem{{ItemName: "Ping"}, {ItemName: "Users"}, {ItemName: "Ping"}, {ItemName: "hml"}})
	if got := changes.UpdateMessage(); got != "Update Ping, Users, hml" {
		t.Fatalf("UpdateMessage = %q", got)
	}
	if got := NewChangeSet([]ChangeItem{{ItemName: "Acme"}}).UpdateMessage(); got != "Update Acme" {
		t.Fatalf("UpdateMessage = %q", got)
	}
}
