package webservice

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func TestGitSyncHelpersOverAClone(t *testing.T) {
	ctx := context.Background()
	servers := gitServers(t)
	for path, content := range map[string]string{"README.md": "root\n", "api/a.yaml": "a\n"} {
		if err := servers.CommitFile("acme/sync", path, content, "seed "+path); err != nil {
			t.Fatal(err)
		}
	}
	git := NewGitWebservice(failingAgent{})
	access := tokenAccess(servers, "acme/sync", "")
	clone := filepath.Join(t.TempDir(), "clone")
	if err := git.Clone(ctx, access, clone, entity.SyncSettings{Branch: "main", Subfolder: "api"}); err != nil {
		t.Fatal(err)
	}
	heads, err := git.ReadHeads(ctx, clone, "main")
	if err != nil || heads.Local == "" || heads.Local != heads.Remote || heads.Ahead != 0 || heads.Behind != 0 {
		t.Fatalf("ReadHeads = %+v, %v", heads, err)
	}
	if err := os.WriteFile(filepath.Join(clone, "api", "b.yaml"), []byte("b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := git.ReadWorktreeFiles(ctx, clone, "api", []string{"a.yaml", "b.yaml"})
	if err != nil || !reflect.DeepEqual(files, map[string][]byte{"a.yaml": []byte("a\n"), "b.yaml": []byte("b\n")}) {
		t.Fatalf("ReadWorktreeFiles = %v, %v", files, err)
	}
	if _, err := git.ReadWorktreeFiles(ctx, clone, "api", []string{"missing.yaml"}); err == nil {
		t.Fatal("a missing worktree file is an error")
	}
	if err := git.StoreWorktreeBlobs(ctx, clone, "api", []string{"b.yaml"}); err != nil {
		t.Fatal(err)
	}
	stored, err := git.ReadBlob(ctx, clone, plumbing.ComputeHash(plumbing.BlobObject, []byte("b\n")).String())
	if err != nil || string(stored) != "b\n" {
		t.Fatalf("stored worktree blob = %q, %v", stored, err)
	}
	expectCode(t, git.StoreWorktreeBlobs(ctx, clone, "api", []string{"missing.yaml"}), "GIT-72500")
	merged, err := git.StoreBlob(ctx, clone, []byte("merged\n"))
	if err != nil || merged != plumbing.ComputeHash(plumbing.BlobObject, []byte("merged\n")).String() {
		t.Fatalf("StoreBlob = %s, %v", merged, err)
	}
	first, err := git.Write(ctx, clone, entity.SyncCommit{Branch: "main", Subfolder: "api", Parent: heads.Local, Message: "first", When: gitTestTime,
		Blobs: map[string]string{"a.yaml": merged, "b.yaml": plumbing.ComputeHash(plumbing.BlobObject, []byte("b\n")).String()}})
	if err != nil {
		t.Fatal(err)
	}
	second, err := git.Write(ctx, clone, entity.SyncCommit{Branch: "main", Subfolder: "api", Parent: first, Message: "second\n\nbody", When: gitTestTime, Files: map[string]string{"c.yaml": "c\n"}})
	if err != nil {
		t.Fatal(err)
	}
	messages, err := git.ListMessages(ctx, clone, heads.Local, second)
	if err != nil || !reflect.DeepEqual(messages, []string{"first", "second\n\nbody"}) {
		t.Fatalf("ListMessages = %q, %v", messages, err)
	}
	ahead, err := git.ReadHeads(ctx, clone, "main")
	if err != nil || ahead.Local != second || ahead.Ahead != 2 || ahead.Behind != 0 || ahead.Base != heads.Local {
		t.Fatalf("ReadHeads after commits = %+v, %v", ahead, err)
	}
	if err := git.MoveBranch(ctx, clone, "main", first); err != nil {
		t.Fatal(err)
	}
	moved, err := git.ReadHeads(ctx, clone, "main")
	if err != nil || moved.Local != first || moved.Ahead != 1 {
		t.Fatalf("ReadHeads after the move = %+v, %v", moved, err)
	}
	if err := servers.CommitFile("acme/sync", "api/z.yaml", "z\n", "teammate"); err != nil {
		t.Fatal(err)
	}
	fetched, err := git.Fetch(ctx, access, clone, "main")
	if err != nil || fetched.Ahead != 1 || fetched.Behind != 1 {
		t.Fatalf("Fetch = %+v, %v", fetched, err)
	}
	all, err := git.ListMessages(ctx, clone, "", first)
	if err != nil || len(all) != 3 || all[2] != "first" {
		t.Fatalf("ListMessages from the root = %q, %v", all, err)
	}
	for _, failure := range []struct {
		err  error
		code string
	}{
		{func() error { _, err := git.ReadHeads(ctx, t.TempDir(), "main"); return err }(), "GIT-63500"},
		{git.MoveBranch(ctx, t.TempDir(), "main", first), "GIT-65500"},
		{func() error { _, err := git.ListMessages(ctx, t.TempDir(), "", first); return err }(), "GIT-67500"},
		{func() error { _, err := git.ListMessages(ctx, clone, "", strings.Repeat("1", 40)); return err }(), "GIT-68500"},
		{git.StoreWorktreeBlobs(ctx, t.TempDir(), "api", nil), "GIT-71500"},
		{func() error { _, err := git.StoreBlob(ctx, t.TempDir(), nil); return err }(), "GIT-74500"},
	} {
		expectCode(t, failure.err, failure.code)
	}
}

func TestGitClaimAllowsOneOperationPerWorkspace(t *testing.T) {
	ctx := context.Background()
	git := NewGitWebservice(failingAgent{})
	release, err := git.Claim(ctx, "w1")
	if err != nil {
		t.Fatal(err)
	}
	_, err = git.Claim(ctx, "w1")
	expectCode(t, err, "GIT-05409")
	if !errs.IsConflictError(err) {
		t.Fatal("busy is a conflict")
	}
	other, err := git.Claim(ctx, "w2")
	if err != nil {
		t.Fatal(err)
	}
	other()
	release()
	again, err := git.Claim(ctx, "w1")
	if err != nil {
		t.Fatal(err)
	}
	again()
}

func TestGitPushErrorsAreClassified(t *testing.T) {
	cases := map[string]struct {
		err      error
		code     string
		status   int
		conflict bool
	}{
		"non fast forward":   {errors.New("non-fast-forward update: refs/heads/main"), "GIT-21409", http.StatusConflict, true},
		"protected branch":   {errors.New("command error on refs/heads/main: protected branch hook declined"), "GIT-14403", http.StatusForbidden, false},
		"signatures":         {errors.New("command error on refs/heads/main: Commits must have verified signatures"), "GIT-14403", http.StatusForbidden, false},
		"github rule":        {errors.New("command error on refs/heads/main: GH006: Protected branch update failed"), "GIT-14403", http.StatusForbidden, false},
		"other rejection":    {errors.New("command error on refs/heads/main: quota exceeded"), "GIT-13409", http.StatusConflict, false},
		"credentials":        {fmt.Errorf("%w: bad", transport.ErrAuthenticationRequired), "GIT-11401", http.StatusUnauthorized, false},
		"organisation":       {fmt.Errorf("%w: the acme organization has enabled OAuth App access restrictions", transport.ErrAuthorizationFailed), "GIT-12403", http.StatusForbidden, false},
		"sso":                {fmt.Errorf("%w: Resource protected by organization SAML enforcement", transport.ErrAuthorizationFailed), "GIT-12403", http.StatusForbidden, false},
		"plain forbidden":    {fmt.Errorf("%w: no", transport.ErrAuthorizationFailed), "GIT-11401", http.StatusUnauthorized, false},
		"missing repository": {transport.ErrRepositoryNotFound, "GIT-15404", http.StatusNotFound, false},
		"unknown":            {errors.New("boom"), "GIT-26500", http.StatusInternalServerError, false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			err := pushError(c.err)
			var base errs.BaseError
			if !errors.As(err, &base) || base.Code() != c.code || base.StatusCode() != c.status || errs.IsConflictError(err) != c.conflict {
				t.Fatalf("pushError = %v (%T)", err, err)
			}
			if c.code != "GIT-26500" && base.ErrorDetails()["attribute"] != nextStepAttribute {
				t.Fatalf("%s carries no next step: %v", c.code, base.ErrorDetails())
			}
		})
	}
}

func TestGitHeadsOfAnEmptyRemoteAndUnrelatedHistories(t *testing.T) {
	ctx := context.Background()
	servers := gitServers(t)
	if err := servers.CreateRemote("acme/fresh"); err != nil {
		t.Fatal(err)
	}
	git := NewGitWebservice(failingAgent{})
	access := tokenAccess(servers, "acme/fresh", "")
	clone := filepath.Join(t.TempDir(), "clone")
	if err := git.Clone(ctx, access, clone, entity.SyncSettings{Branch: "main", Subfolder: "api"}); err != nil {
		t.Fatal(err)
	}
	local, err := git.Write(ctx, clone, entity.SyncCommit{Branch: "main", Subfolder: "api", Message: "mine", When: gitTestTime, Files: map[string]string{"a.yaml": "a\n"}})
	if err != nil {
		t.Fatal(err)
	}
	versions, err := git.ReadVersions(ctx, clone, "api", entity.BranchHeads{Local: local})
	if err != nil || len(versions.Mine) != 1 {
		t.Fatalf("the committed subfolder = %+v, %v", versions, err)
	}
	heads, err := git.ReadHeads(ctx, clone, "main")
	if err != nil || heads.Local != local || heads.Remote != "" || heads.Base != "" || heads.Ahead != 1 || heads.Behind != 0 {
		t.Fatalf("heads before any remote commit = %+v, %v", heads, err)
	}
	if err := servers.CommitFile("acme/fresh", "README.md", "theirs\n", "theirs"); err != nil {
		t.Fatal(err)
	}
	unrelated, err := git.Fetch(ctx, access, clone, "main")
	if err != nil || unrelated.Base != "" || unrelated.Ahead != 1 || unrelated.Behind != 1 {
		t.Fatalf("unrelated histories = %+v, %v", unrelated, err)
	}
}
