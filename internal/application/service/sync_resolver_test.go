package service

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func TestTheirCopyPathIsUniqueNextToTheConflict(t *testing.T) {
	conflict := entity.ConflictItem{Path: "collections/api/users.yaml", ItemName: "Users", ItemKind: itemkind.Request}
	if got := theirCopyPath(conflict, map[string]string{"collections/api/users.yaml": "1"}); got != "collections/api/users-(theirs).yaml" {
		t.Fatalf("copy path %q", got)
	}
	taken := map[string]string{"collections/api/users-(theirs).yaml": "1", "collections/other/users-(theirs)-2.yaml": "2"}
	worktree := map[string]string{"collections/api/users-(theirs)-2.yaml": "3"}
	if got := theirCopyPath(conflict, taken, worktree); got != "collections/api/users-(theirs)-3.yaml" {
		t.Fatalf("copy path %q", got)
	}
	environment := entity.ConflictItem{Path: "environments/dev.yaml", ItemName: "Dev", ItemKind: itemkind.Environment}
	if got := theirCopyPath(environment); got != "environments/dev-(theirs).yaml" {
		t.Fatalf("environment copy %q", got)
	}
	root := entity.ConflictItem{Path: "ping.yaml", ItemName: "Ping", ItemKind: itemkind.Request}
	if got := theirCopyPath(root, map[string]string{"api/ping-(theirs).yaml": "1"}); got != "ping-(theirs).yaml" {
		t.Fatalf("root copy %q", got)
	}
}
