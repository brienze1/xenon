package persistence

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/domain/enums/nodekind"
)

const syncedRequest = "xenon: 1\nkind: request\nid: 01JS0000000000000000000002\nname: Ping\nmethod: GET\nurl: https://a.test\n"

func syncedWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "workspace.yaml"), "xenon: 1\nkind: workspace\nid: 01JS0000000000000000000000\nname: Acme\nsync: {}\n")
	writeTestFile(t, filepath.Join(root, "collections", "api", "collection.yaml"), "xenon: 1\nkind: collection\nid: 01JS0000000000000000000001\nname: API\n")
	writeTestFile(t, filepath.Join(root, "collections", "api", "ping.yaml"), syncedRequest)
	writeTestFile(t, filepath.Join(root, "collections", "api", "old", "folder.yaml"), "xenon: 1\nkind: folder\nid: 01JS0000000000000000000005\nname: Old\n")
	writeTestFile(t, filepath.Join(root, "collections", "api", "old", "gone.yaml"), strings.ReplaceAll(syncedRequest, "0002\nname: Ping", "0006\nname: Gone"))
	return root
}

func TestApplySyncedFilesWritesOwnFilesAndReportsRequests(t *testing.T) {
	ctx := context.Background()
	index := NewWorkspaceIndexCache()
	tree := NewTreePersistence(index)
	root := syncedWorkspace(t)
	changed := strings.Replace(syncedRequest, "GET", "POST", 1)
	added := strings.ReplaceAll(syncedRequest, "0002\nname: Ping", "0003\nname: Users")
	changes, err := tree.ApplySyncedFiles(ctx, root, map[string][]byte{
		"collections/api/ping.yaml":  []byte(changed),
		"collections/api/users.yaml": []byte(added),
	}, []string{"collections/api/old/gone.yaml", "collections/api/old/folder.yaml", "collections/api/never.yaml"})
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(changes))
	for _, change := range changes {
		ids = append(ids, change.RequestID)
		if (change.RequestID == "01JS0000000000000000000006") != change.Deleted {
			t.Errorf("change %+v", change)
		}
	}
	if !reflect.DeepEqual(ids, []string{"01JS0000000000000000000002", "01JS0000000000000000000003", "01JS0000000000000000000006"}) {
		t.Fatalf("ids = %v", ids)
	}
	if content, _ := os.ReadFile(filepath.Join(root, "collections", "api", "ping.yaml")); string(content) != changed {
		t.Fatalf("ping = %q", content)
	}
	if _, err := os.Stat(filepath.Join(root, "collections", "api", "old")); !os.IsNotExist(err) {
		t.Fatal("emptied folders are removed")
	}
	if _, err := os.Stat(filepath.Join(root, "collections", "api")); err != nil {
		t.Fatal("the collection stays")
	}
	for _, path := range []string{filepath.Join(root, "collections", "api", "ping.yaml"), filepath.Join(root, "collections", "api", "users.yaml")} {
		data, _ := os.ReadFile(path)
		if !index.IsOwnWrite(ctx, path, data) {
			t.Errorf("%s is not an own write", path)
		}
	}
	if !index.IsOwnMove(ctx, filepath.Join(root, "collections", "api", "old", "gone.yaml")) {
		t.Error("a pulled deletion is an own change")
	}
	writeTestFile(t, filepath.Join(root, "notes.md"), "x")
	if _, err := tree.ApplySyncedFiles(ctx, root, nil, []string{"notes.md", "collections/api/ping.yaml", "collections/api/users.yaml"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "collections", "api")); err != nil {
		t.Fatal("a collection keeps its folder while collection.yaml is there")
	}
	writeTestFile(t, filepath.Join(root, "environments", "solo", "deep", "dev.yaml"), "x")
	if _, err := tree.ApplySyncedFiles(ctx, root, nil, []string{"environments/solo/deep/dev.yaml"}); err != nil {
		t.Fatal(err)
	}
	expectOnly(t, filepath.Join(root, "environments"))
	expectOnly(t, root, "collections", "environments", "workspace.yaml")
	if _, err := tree.ApplySyncedFiles(ctx, root, map[string][]byte{"collections/api/collection.yaml/inside.yaml": []byte("x")}, nil); errorCode(err) != "GIT-76500" {
		t.Fatalf("writing under a file = %v", err)
	}
}

func TestIdentifyItemsReadsTheItemHeader(t *testing.T) {
	refs := NewTreePersistence(NewWorkspaceIndexCache()).IdentifyItems(context.Background(), map[string][]byte{
		"collections/api/ping.yaml": []byte(syncedRequest),
		"workspace.yaml":            []byte("kind: workspace\nid: w1\nname: Acme\n"),
		"broken.yaml":               []byte(": not yaml ["),
		"no-id.yaml":                []byte("kind: request\nname: x\n"),
		"odd-kind.yaml":             []byte("kind: banana\nid: b\n"),
		".gitignore":                []byte(".xenon/\n"),
	})
	want := map[string]entity.ItemRef{
		"collections/api/ping.yaml": {ID: "01JS0000000000000000000002", Name: "Ping", Kind: itemkind.Request},
		"workspace.yaml":            {ID: "w1", Name: "Acme", Kind: itemkind.Workspace},
	}
	if !reflect.DeepEqual(refs, want) {
		t.Fatalf("refs = %+v", refs)
	}
}

func TestSyncDocumentsRoundTrip(t *testing.T) {
	ctx := context.Background()
	tree := NewTreePersistence(NewWorkspaceIndexCache())
	source := "xenon: 1\nkind: collection\nname: API\ndescription: |-\n  line one\n  line two\nempty: {}\nnone: []\norder:\n  - b.yaml\n  - a.yaml\nanchor: &x hello\nalias: *x\n"
	node, err := tree.DecodeDocument(ctx, []byte(source))
	if err != nil {
		t.Fatal(err)
	}
	if node.Kind != nodekind.Mapping || node.Field("alias").Value != "hello" || node.Field("xenon").Tag != "!!int" || len(node.Field("order").Items) != 2 {
		t.Fatalf("node = %+v", node)
	}
	encoded, err := tree.EncodeDocument(ctx, node)
	if err != nil {
		t.Fatal(err)
	}
	want := "xenon: 1\nkind: collection\nname: API\ndescription: |-\n  line one\n  line two\nempty: {}\nnone: []\norder:\n  - b.yaml\n  - a.yaml\nanchor: hello\nalias: hello\n"
	if string(encoded) != want {
		t.Fatalf("encoded =\n%s", encoded)
	}
	nested := entity.NewMappingNode().Set("missing", nil)
	if encoded, err := tree.EncodeDocument(ctx, nested); err != nil || string(encoded) != "missing: null\n" {
		t.Fatalf("nil field = %q, %v", encoded, err)
	}
	if empty, err := tree.DecodeDocument(ctx, nil); err != nil || !empty.IsMapping() {
		t.Fatalf("empty document = %+v, %v", empty, err)
	}
	if _, err := tree.DecodeDocument(ctx, []byte("- a\n- b\n")); errorCode(err) != "GIT-79422" {
		t.Fatalf("a list document = %v", err)
	}
}

func TestWorkspaceSyncSectionCopiesAndClearing(t *testing.T) {
	ctx := context.Background()
	workspaces := NewWorkspacePersistence(NewWorkspaceIndexCache())
	root := syncedWorkspace(t)
	settings := entity.NewSyncSettings("https://git.test/acme.git", "main", "api/xenon")
	if err := workspaces.SaveSync(ctx, root, settings); err != nil {
		t.Fatal(err)
	}
	loaded, err := workspaces.Load(ctx, root)
	if err != nil || loaded.Sync != settings || loaded.Name != "Acme" {
		t.Fatalf("Load = %+v, %v", loaded, err)
	}
	content, _ := os.ReadFile(filepath.Join(root, "workspace.yaml"))
	if !strings.Contains(string(content), "sync:\n  repository: https://git.test/acme.git\n  branch: main\n  subfolder: api/xenon\n") {
		t.Fatalf("workspace.yaml =\n%s", content)
	}
	if err := workspaces.SaveSync(ctx, filepath.Join(root, "missing"), settings); errorCode(err) != "WKS-06404" {
		t.Fatalf("missing workspace = %v", err)
	}
	clone := t.TempDir()
	writeTestFile(t, filepath.Join(clone, ".git", "HEAD"), "ref: refs/heads/main\n")
	writeTestFile(t, filepath.Join(clone, "README.md"), "root\n")
	writeTestFile(t, filepath.Join(root, ".xenon", "local.json"), "{}")
	if err := workspaces.CopyInto(ctx, root, clone); err != nil {
		t.Fatal(err)
	}
	expectOnly(t, clone, ".git", "README.md", "collections", "workspace.yaml")
	if err := workspaces.ClearFolder(ctx, clone); err != nil {
		t.Fatal(err)
	}
	expectOnly(t, clone, ".git")
	if err := workspaces.ClearFolder(ctx, filepath.Join(clone, "missing")); err != nil {
		t.Fatal(err)
	}
	if err := workspaces.CopyInto(ctx, filepath.Join(root, "missing"), clone); errorCode(err) != "GIT-83500" {
		t.Fatalf("missing source = %v", err)
	}
}

func TestLocalSyncSettingsAndTheDefaultClone(t *testing.T) {
	ctx := context.Background()
	dataDir := t.TempDir()
	t.Setenv("XENON_DATA_DIR", dataDir)
	state := NewUIStatePersistence()
	settings := entity.NewLocalSyncSettings(entity.NewSyncSettings("r", "main", "api"), "acc", filepath.Join(dataDir, "c"))
	if err := state.SaveSyncSettings(ctx, "w1", settings.WithOptions("acc", true, false)); err != nil {
		t.Fatal(err)
	}
	found, err := state.FindSyncSettings(ctx, "w1")
	if err != nil || *found != settings.WithOptions("acc", true, false) {
		t.Fatalf("FindSyncSettings = %+v, %v", found, err)
	}
	if clone := state.DefaultClonePath(ctx, "w1"); clone != filepath.Join(dataDir, "repos", "w1") {
		t.Fatalf("DefaultClonePath = %s", clone)
	}
}

func TestSavingTheRegistryKeepsTheGitAccounts(t *testing.T) {
	ctx := context.Background()
	dataDir := t.TempDir()
	t.Setenv("XENON_DATA_DIR", dataDir)
	writeTestFile(t, filepath.Join(dataDir, "app.json"), `{"version": 1, "workspaces": [], "gitAccounts": [{"id": "a1"}]}`)
	registry := NewRegistryPersistence()
	stored, err := registry.Find(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.Save(ctx, *stored); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dataDir, "app.json"))
	var saved map[string]json.RawMessage
	if err := json.Unmarshal(data, &saved); err != nil || !strings.Contains(string(saved["gitAccounts"]), `"a1"`) {
		t.Fatalf("app.json = %s", data)
	}
}
