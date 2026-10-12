package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func TestSyncSessionsAreKeptPerWorkspace(t *testing.T) {
	ctx := context.Background()
	sessions := NewSyncSessionPersistence()
	if found, err := sessions.FindWorkspaceSession(ctx, "w1"); found != nil || err != nil {
		t.Fatalf("no session yet: %v %v", found, err)
	}
	first := entity.NewSyncSession("s1", "w1", entity.BranchHeads{Local: "l"}, true, nil, time.Now())
	if err := sessions.SaveSession(ctx, first); err != nil {
		t.Fatal(err)
	}
	if found, err := sessions.FindSession(ctx, "s1"); err != nil || found.LocalID != "w1" {
		t.Fatalf("find by id: %+v %v", found, err)
	}
	replaced := entity.NewSyncSession("s2", "w1", entity.BranchHeads{Local: "l"}, true, nil, time.Now())
	if err := sessions.SaveSession(ctx, replaced); err != nil {
		t.Fatal(err)
	}
	if _, err := sessions.FindSession(ctx, "s1"); errorCode(err) != "GIT-16404" {
		t.Fatalf("a replaced session is gone: %v", err)
	}
	if err := sessions.UpdateSession(ctx, first); errorCode(err) != "GIT-38404" {
		t.Fatalf("updating a replaced session: %v", err)
	}
	if err := sessions.UpdateSession(ctx, entity.NewSyncSession("s3", "w2", entity.BranchHeads{}, false, nil, time.Now())); errorCode(err) != "GIT-38404" {
		t.Fatalf("updating a missing session: %v", err)
	}
	updated := replaced
	updated.Replay = false
	if err := sessions.UpdateSession(ctx, updated); err != nil {
		t.Fatal(err)
	}
	if found, _ := sessions.FindWorkspaceSession(ctx, "w1"); found == nil || found.ID != "s2" || found.Replay {
		t.Fatalf("updated session %+v", found)
	}
	if err := sessions.DiscardSession(ctx, "w1"); err != nil {
		t.Fatal(err)
	}
	if found, _ := sessions.FindWorkspaceSession(ctx, "w1"); found != nil {
		t.Fatal("discarded")
	}
	if err := sessions.DiscardSession(ctx, "w1"); err != nil {
		t.Fatal("discarding twice is fine")
	}
}

func TestRenderItemParsesRequestsAndEnvironments(t *testing.T) {
	ctx := context.Background()
	tree := NewTreePersistence(NewWorkspaceIndexCache())
	request := tree.RenderItem(ctx, itemkind.Request, "collections/api/ping.yaml", []byte(syncedRequest))
	if request.Request == nil || request.Request.ID != "01JS0000000000000000000002" || request.Request.Method != "GET" || request.Text != syncedRequest || request.Environment != nil {
		t.Fatalf("request version %+v", request)
	}
	environmentYaml := "xenon: 1\nkind: environment\nid: 01JS0000000000000000000004\nname: Dev\nvariables:\n  - key: token\n    secret: true\n  - key: base\n    value: https://a.test\n"
	environment := tree.RenderItem(ctx, itemkind.Environment, "environments/dev.yaml", []byte(environmentYaml))
	if environment.Environment == nil || environment.Environment.Name != "Dev" || environment.Environment.FileName != "dev.yaml" || len(environment.Environment.Variables) != 2 || environment.Request != nil {
		t.Fatalf("environment version %+v", environment)
	}
	broken := tree.RenderItem(ctx, itemkind.Request, "x.yaml", []byte("kind: [unclosed"))
	if broken.Request != nil || broken.Text != "kind: [unclosed" {
		t.Fatalf("broken request %+v", broken)
	}
	wrongKind := tree.RenderItem(ctx, itemkind.Environment, "environments/x.yaml", []byte(syncedRequest))
	if wrongKind.Environment != nil {
		t.Fatal("a request is not an environment")
	}
	folder := tree.RenderItem(ctx, itemkind.Folder, "a/folder.yaml", []byte("name: A\n"))
	if folder.Request != nil || folder.Environment != nil || folder.Text != "name: A\n" {
		t.Fatalf("folders render as text only %+v", folder)
	}
}

func TestRenderNodeShowsScalarsAsValuesAndTreesAsYaml(t *testing.T) {
	ctx := context.Background()
	tree := NewTreePersistence(NewWorkspaceIndexCache())
	if text, err := tree.RenderNode(ctx, nil); text != nil || err != nil {
		t.Fatal("absent node")
	}
	if text, err := tree.RenderNode(ctx, entity.NewScalarNode("!!str", "Mine")); err != nil || *text != "Mine" {
		t.Fatalf("scalar %v %v", text, err)
	}
	mapping := entity.NewMappingNode().Set("type", entity.NewScalarNode("!!str", "bearer")).Set("token", entity.NewScalarNode("!!str", "{{t}}"))
	if text, err := tree.RenderNode(ctx, mapping); err != nil || *text != "type: bearer\ntoken: '{{t}}'" {
		t.Fatalf("mapping %q %v", *text, err)
	}
}
