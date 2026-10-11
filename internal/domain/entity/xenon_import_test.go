package entity

import (
	"reflect"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/importitemkind"
	"github.com/brienze1/xenon/internal/domain/enums/importstatus"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/domain/enums/vaultstate"
)

func TestImportProbeXenon(t *testing.T) {
	cases := []struct {
		probe                         ImportProbe
		xenon, newer, bundle, exports bool
	}{
		{ImportProbe{XenonFormat: 1, XenonKind: "request"}, true, false, false, false},
		{ImportProbe{XenonFormat: 0, XenonKind: "request"}, false, false, false, false},
		{ImportProbe{XenonFormat: 1}, false, false, false, false},
		{ImportProbe{XenonFormat: 2, XenonKind: "bundle"}, true, true, true, false},
		{ImportProbe{XenonFormat: 1, XenonKind: "export"}, true, false, false, true},
	}
	for _, tc := range cases {
		if tc.probe.IsXenon() != tc.xenon || tc.probe.IsNewerXenon() != tc.newer || tc.probe.IsXenonBundle() != tc.bundle || tc.probe.IsXenonExport() != tc.exports {
			t.Errorf("%+v", tc.probe)
		}
	}
}

func TestImportNodeReidentified(t *testing.T) {
	request := NewRequest("r1", "R")
	request.Examples = []SavedExample{NewSavedExample("e1", "E")}
	bare := NewRequest("r2", "Bare")
	folder := NewContainerImportNode(Collection{ID: "f1", Kind: itemkind.Folder, Name: "F"}, []ImportNode{NewRequestImportNode(request), NewRequestImportNode(bare)})
	root := NewContainerImportNode(Collection{ID: "c1", Kind: itemkind.Collection, Name: "C"}, []ImportNode{folder})
	next := 0
	renewed := root.Reidentified(func() string {
		next++
		return "n" + string(rune('0'+next))
	})
	if renewed.ID() != "n1" || renewed.Children[0].ID() != "n2" || renewed.Children[0].Children[0].ID() != "n3" || renewed.Children[0].Children[0].Request.Examples[0].ID != "n4" || renewed.Children[0].Children[1].ID() != "n5" {
		t.Errorf("ids = %v", renewed.IDs())
	}
	if renewed.Children[0].Children[1].Request.Examples != nil || renewed.Children[0].Children[0].Request.Name != "R" {
		t.Error("a request without examples keeps none and names are kept")
	}
	if root.ID() != "c1" || request.Examples[0].ID != "e1" {
		t.Error("the original is not changed")
	}
}

func TestImportItemKindOfAndPlacement(t *testing.T) {
	request := NewRequestImportNode(NewRequest("r", "R"))
	folder := NewContainerImportNode(Collection{ID: "f", Kind: itemkind.Folder, Name: "F"}, []ImportNode{request})
	collection := NewContainerImportNode(Collection{ID: "c", Kind: itemkind.Collection, Name: "C"}, []ImportNode{request})
	if ImportItemKindOf(request) != importitemkind.Request || ImportItemKindOf(folder) != importitemkind.Folder || ImportItemKindOf(collection) != importitemkind.Collection {
		t.Error("item kinds")
	}
	if top := NewTopLevelPlacement(folder); top.Nodes[0].Container.Kind != itemkind.Collection || folder.Container.Kind != itemkind.Folder {
		t.Errorf("a top-level folder becomes a collection: %+v", top.Nodes[0].Container.Kind)
	}
	if top := NewTopLevelPlacement(request); top.Nodes[0].Kind() != itemkind.Request {
		t.Error("a request stays a request")
	}
	parent := Collection{ID: "p", Kind: itemkind.Collection, Name: "P"}
	placed := folder.PlacedInto(parent)
	if placed.Merge || placed.Wrapped || len(placed.Nodes) != 1 || placed.Nodes[0].Container.Kind != itemkind.Folder || placed.Parent.ID != "p" {
		t.Errorf("a folder goes in as a folder: %+v", placed)
	}
	if wrapped := collection.PlacedInto(parent); !wrapped.Merge && !wrapped.Wrapped {
		t.Error("a collection into a parent is still merged or wrapped")
	}
}

func TestImportItemSecrets(t *testing.T) {
	bundle := NewSecretBundle(map[string]map[string]string{"env:e-old": {"key": "v", "plain": "p"}, "collection:c-old": {"token": "t"}, "globals": {"g": "1"}})
	environment := Environment{ID: "e-new", Name: "local", Variables: []VariableRow{{Key: "key", Secret: true}, NewVariableRow("plain", "x")}}
	variables := NewVariablesImportItem("0", importitemkind.Environment, NewImportVariables(environment, nil)).FromSource("e-old").WithSecrets(bundle)
	if variables.SourceID != "e-old" || !reflect.DeepEqual(variables.Variables.SecretValues(), map[string]string{"key": "v"}) {
		t.Errorf("environment secrets = %v", variables.Variables.SecretValues())
	}
	tree := NewImportItem("1", importitemkind.Collection, NewContainerImportNode(Collection{ID: "c-new"}, nil), nil).FromSource("c-old").WithSecrets(bundle)
	if !reflect.DeepEqual(tree.CollectionSecrets, map[string]string{"token": "t"}) {
		t.Errorf("collection secrets = %v", tree.CollectionSecrets)
	}
	workspace := NewWorkspaceImportItem("2", ImportWorkspace{Name: "Acme", Counts: ImportCounts{Requests: 3}}).WithSecrets(bundle)
	if !workspace.IsWorkspace() || workspace.Kind != importitemkind.Workspace || workspace.Name != "Acme" || len(workspace.Workspace.Secrets.Scopes) != 3 || workspace.Counts().Requests != 3 {
		t.Errorf("workspace = %+v", workspace)
	}
	if len(workspace.Scripts()) != 0 || tree.IsWorkspace() || variables.IsWorkspace() {
		t.Error("only the workspace item is a workspace")
	}
	outcome := workspace.CompletedWorkspace("local-1", "ws-1")
	if outcome.Item.Status != importstatus.Imported || outcome.Item.LocalID != "local-1" || outcome.Item.ID != "ws-1" || outcome.Item.Name != "Acme" || outcome.Item.Key != "2" || outcome.Created.Requests != 3 || outcome.Warnings == nil || outcome.PmUsages == nil {
		t.Errorf("outcome = %+v", outcome)
	}
}

func TestImportSourceSecrets(t *testing.T) {
	source := NewImportSource("s", nil, nil, time.Now()).WithItems([]ImportItem{
		NewImportItem("0", importitemkind.Collection, NewContainerImportNode(Collection{ID: "c-new"}, nil), nil).FromSource("c-old"),
	})
	if source.HasEncryptedSecrets() || !source.WithEncryptedSecrets("armored").HasEncryptedSecrets() {
		t.Error("encrypted secrets flag")
	}
	unlocked := source.WithSecrets(NewSecretBundle(map[string]map[string]string{"collection:c-old": {"k": "v"}}))
	if unlocked.Items[0].CollectionSecrets["k"] != "v" || source.Items[0].CollectionSecrets != nil {
		t.Error("secrets go to the matching item of a copy")
	}
	plan := ImportPlan{}
	if !plan.NeedsPassphrase(source.WithEncryptedSecrets("armored")) || plan.NeedsPassphrase(source) || (ImportPlan{SkipSecrets: true}).NeedsPassphrase(source.WithEncryptedSecrets("armored")) {
		t.Error("a passphrase is needed for encrypted secrets unless they are skipped")
	}
	if logged, _ := (ImportPlan{Passphrase: "p", SkipSecrets: true}).LogValue().(map[string]any); logged["passphrase"] != nil || logged["skipSecrets"] != true {
		t.Errorf("plan log = %v", logged)
	}
}

func TestImportWorkspaceCounts(t *testing.T) {
	request := NewRequest("r", "R")
	request.Examples = []SavedExample{NewSavedExample("e", "E")}
	folder := NewContainerNode(Collection{ID: "f", Kind: itemkind.Folder}, "f", nil)
	folder.Children = []TreeNode{NewRequestNode(request, "r.yaml"), NewErrorNode("x", "x.yaml", "broken", false)}
	collection := NewContainerNode(Collection{ID: "c", Kind: itemkind.Collection}, "c", nil)
	collection.Children = []TreeNode{folder, NewRequestNode(NewRequest("q", "Q"), "q.yaml")}
	environments := []Environment{{ID: "e", Variables: []VariableRow{NewVariableRow("a", "b")}}, NewMalformedEnvironment("bad.yaml", "broken")}
	workspace := NewImportWorkspace("/w", "Acme", []TreeNode{collection, NewErrorNode("y", "y", "ignored", true)}, environments)
	want := ImportCounts{Collections: 1, Folders: 1, Requests: 2, Examples: 1, Environments: 1, Variables: 1}
	if workspace.Counts != want || workspace.Folder != "/w" || workspace.Name != "Acme" || !workspace.Secrets.IsEmpty() {
		t.Errorf("workspace = %+v", workspace)
	}
	if logged, _ := workspace.LogValue().(map[string]any); logged["name"] != "Acme" || logged["folder"] != "/w" {
		t.Errorf("log = %v", logged)
	}
}

func TestXenonImportLogValue(t *testing.T) {
	root := NewRequestImportNode(NewRequest("r", "R"))
	environment := Environment{Name: "local"}
	workspace := ImportWorkspace{Name: "Acme"}
	logged, _ := (XenonImport{Root: &root, Environment: &environment, Workspace: &workspace, Secrets: "armored"}).LogValue().(map[string]any)
	if logged["root"] != "R" || logged["environment"] != "local" || logged["workspace"] == nil || logged["hasSecrets"] != true || logged["secrets"] != nil {
		t.Errorf("log = %v", logged)
	}
	empty, _ := (XenonImport{}).LogValue().(map[string]any)
	if len(empty) != 1 || empty["hasSecrets"] != false {
		t.Errorf("empty log = %v", empty)
	}
}

func TestImportResultLocalIDsSkipsFailedWorkspaces(t *testing.T) {
	result := NewImportResult().
		WithOutcome(ImportOutcome{Item: ImportItemResult{LocalID: "a"}}).
		WithOutcome(ImportOutcome{Item: ImportItemResult{LocalID: ""}, Failure: &ImportFailure{}})
	if ids := result.LocalIDs(); !reflect.DeepEqual(ids, []string{"a"}) {
		t.Errorf("ids = %v", ids)
	}
}

func TestVaultWithBundle(t *testing.T) {
	vault := NewVault("l", "", vaultstate.Unlocked, map[string]map[string]string{"globals": {"old": "1"}})
	updated := vault.WithBundle(NewSecretBundle(map[string]map[string]string{"env:e": {"k": "v"}, "globals": {"g": "2"}}))
	if updated.SecretsFor(NewEnvironmentScopeRef("e"))["k"] != "v" || !reflect.DeepEqual(updated.SecretsFor(NewGlobalsScopeRef()), map[string]string{"g": "2"}) {
		t.Errorf("vault = %+v", updated.Secrets)
	}
	if vault.SecretsFor(NewGlobalsScopeRef())["old"] != "1" {
		t.Error("the original vault is not changed")
	}
}

func TestImportResultLeavesOutWorkspacelessFailures(t *testing.T) {
	failure := ImportFailure{ItemKey: "0", Name: "Acme"}
	result := NewImportResult().WithOutcome(ImportOutcome{Item: ImportItemResult{Key: "0", Name: "Acme"}, Failure: &failure})
	if len(result.Workspaces) != 0 || len(result.Failed) != 1 || len(result.Items) != 1 {
		t.Errorf("result = %+v", result)
	}
	result = result.WithOutcome(ImportOutcome{Item: ImportItemResult{Key: "1", LocalID: "local"}})
	if len(result.Workspaces) != 1 || result.Workspaces[0].LocalID != "local" {
		t.Errorf("workspaces = %+v", result.Workspaces)
	}
}

func TestImportPredicates(t *testing.T) {
	workspace := XenonImport{Workspace: &ImportWorkspace{}}
	environment := XenonImport{Environment: &Environment{}}
	if !workspace.IsWorkspace() || workspace.IsEnvironment() || !environment.IsEnvironment() || environment.IsWorkspace() {
		t.Error("xenon import kinds are told apart")
	}
	if (ImportPlan{}).HasPassphrase() || !(ImportPlan{Passphrase: "secret"}).HasPassphrase() {
		t.Error("a plan has a passphrase only when one was typed")
	}
	if (ImportPlanItem{}).HasTargetFolder() || !(ImportPlanItem{TargetFolder: "/w"}).HasTargetFolder() {
		t.Error("a plan item has a target folder only when one was chosen")
	}
	if (ImportItem{}).HasCollectionSecrets() || !(ImportItem{CollectionSecrets: map[string]string{"k": "v"}}).HasCollectionSecrets() {
		t.Error("an item has collection secrets only when some were unlocked")
	}
}
