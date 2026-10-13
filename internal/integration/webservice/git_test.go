package webservice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/gitprovider"
	"github.com/brienze1/xenon/internal/domain/enums/sshmode"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/test/integration/mock"
)

const gitTestToken = "test-token"

var gitTestTime = time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)

type failingAgent struct{}

func (failingAgent) Signers() ([]gossh.Signer, io.Closer, error) {
	return nil, nil, errors.New("no agent")
}

func gitServers(t *testing.T) *mock.GitServers {
	t.Helper()
	servers := mock.NewGitServers()
	if err := servers.Reset(); err != nil {
		t.Fatal(err)
	}
	return servers
}

func tokenAccess(servers *mock.GitServers, name, token string) entity.GitAccess {
	account := entity.NewTokenAccount("a1", gitprovider.GitLab, servers.HttpHost(), "luis", gitTestTime)
	return entity.NewGitAccess(servers.HttpURL(name), account, token)
}

func expectCode(t *testing.T, err error, code string) {
	t.Helper()
	var base errs.BaseError
	if !errors.As(err, &base) || base.Code() != code {
		t.Fatalf("error = %v, want code %s", err, code)
	}
}

func describeError(err error) string {
	var base errs.BaseError
	if errors.As(err, &base) {
		return fmt.Sprintf("%s %s (cause: %v)", base.Code(), base.Description(), errors.Unwrap(err))
	}
	return err.Error()
}

func TestGitCloneCommitPushAndFetchOverHttp(t *testing.T) {
	ctx := context.Background()
	servers := gitServers(t)
	servers.RequireHttpCredentials("oauth2", gitTestToken, false)
	for path, content := range map[string]string{"README.md": "root\n", "api/xenon/workspace.yaml": "name: Acme\n", "api/xenon/collections/a/r.yaml": "r\n", "web/app.js": "x\n"} {
		if err := servers.CommitFile("acme/backend", path, content, "seed "+path); err != nil {
			t.Fatal(err)
		}
	}
	git := NewGitWebservice(failingAgent{})
	access := tokenAccess(servers, "acme/backend", gitTestToken)
	info, err := git.List(ctx, access)
	if err != nil || info.DefaultBranch != "main" || !reflect.DeepEqual(info.Branches, []string{"main"}) {
		t.Fatalf("List = %+v, %v", info, err)
	}
	clone := filepath.Join(t.TempDir(), "clone")
	settings := entity.SyncSettings{Repository: access.Repository, Branch: "main", Subfolder: "api/xenon"}
	if err := git.Clone(ctx, access, clone, settings); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(clone, "web")); !os.IsNotExist(err) {
		t.Fatalf("only the subfolder may be checked out: %v", err)
	}
	worktree, err := git.ReadWorktree(ctx, clone, "api/xenon")
	if err != nil || len(worktree) != 2 || worktree["collections/a/r.yaml"] != plumbing.ComputeHash(plumbing.BlobObject, []byte("r\n")).String() {
		t.Fatalf("ReadWorktree = %v, %v", worktree, err)
	}
	heads, err := git.Fetch(ctx, access, clone, "main")
	if err != nil || heads.Local == "" || heads.Local != heads.Remote || heads.Base != heads.Local {
		t.Fatalf("Fetch = %+v, %v", heads, err)
	}
	commit := entity.SyncCommit{Branch: "main", Subfolder: "api/xenon", Parent: heads.Local, Message: "Update", AuthorName: "Luis", AuthorEmail: "luis@acme.test", When: gitTestTime,
		Files: map[string]string{"workspace.yaml": "name: Acme 2\n", "environments/dev.yaml": "name: dev\n"}}
	written, err := git.Write(ctx, clone, commit)
	if err != nil {
		t.Fatal(err)
	}
	if err := git.Push(ctx, access, clone, "main", heads.Remote); err != nil {
		t.Fatal(err)
	}
	last, err := servers.LastCommit("acme/backend", "main")
	if err != nil || last.Hash.String() != written || last.Author.Email != "luis@acme.test" || last.Message != "Update" {
		t.Fatalf("remote head %v, %v", last, err)
	}
	for path, want := range map[string]string{"README.md": "root\n", "web/app.js": "x\n", "api/xenon/workspace.yaml": "name: Acme 2\n", "api/xenon/environments/dev.yaml": "name: dev\n"} {
		if got, err := servers.ReadFile("acme/backend", "main", path); err != nil || got != want {
			t.Errorf("remote %s = %q, %v", path, got, err)
		}
	}
	if _, err := servers.ReadFile("acme/backend", "main", "api/xenon/collections/a/r.yaml"); err == nil {
		t.Error("files left out of the commit must be deleted")
	}
	if err := servers.CommitFile("acme/backend", "api/xenon/workspace.yaml", "name: Theirs\n", "teammate"); err != nil {
		t.Fatal(err)
	}
	moved, err := git.Fetch(ctx, access, clone, "main")
	if err != nil || moved.Local != written || moved.Base != written || moved.Remote == written {
		t.Fatalf("Fetch after a teammate push = %+v, %v", moved, err)
	}
	versions, err := git.ReadVersions(ctx, clone, "api/xenon", *moved)
	if err != nil || versions.Mine["workspace.yaml"] == versions.Theirs["workspace.yaml"] || versions.Base["environments/dev.yaml"] == "" {
		t.Fatalf("ReadVersions = %+v, %v", versions, err)
	}
	content, err := git.ReadBlob(ctx, clone, versions.Theirs["workspace.yaml"])
	if err != nil || string(content) != "name: Theirs\n" {
		t.Fatalf("ReadBlob = %q, %v", content, err)
	}
	ahead, err := git.Write(ctx, clone, entity.SyncCommit{Branch: "main", Subfolder: "api/xenon", Parent: moved.Local, Message: "stale", When: gitTestTime, Files: map[string]string{"workspace.yaml": "x\n"}})
	if err != nil || ahead == "" {
		t.Fatal(err)
	}
	expectCode(t, git.Push(ctx, access, clone, "main", moved.Local), "GIT-21409")
	if err := git.Checkout(ctx, clone, "main", "api/xenon"); err != nil {
		t.Fatal(err)
	}
	checkedOut, _ := os.ReadFile(filepath.Join(clone, "api", "xenon", "workspace.yaml"))
	if string(checkedOut) != "x\n" {
		t.Fatalf("Checkout wrote %q", checkedOut)
	}
}

func TestGitHttpFailuresCarryTheirCodes(t *testing.T) {
	ctx := context.Background()
	servers := gitServers(t)
	servers.RequireHttpCredentials("oauth2", gitTestToken, false)
	if err := servers.CommitFile("acme/backend", "a.txt", "a", "seed"); err != nil {
		t.Fatal(err)
	}
	git := NewGitWebservice(failingAgent{})
	_, err := git.List(ctx, tokenAccess(servers, "acme/backend", "wrong"))
	expectCode(t, err, "GIT-11401")
	_, err = git.Verify(ctx, tokenAccess(servers, "acme/backend", "wrong"))
	expectCode(t, err, "ACC-02401")
	var base errs.BaseError
	if errors.As(err, &base); base.StatusCode() != 401 || base.ErrorDetails()["attribute"] != nextStepAttribute {
		t.Fatalf("status %d details %v", base.StatusCode(), base.ErrorDetails())
	}
	_, err = git.List(ctx, tokenAccess(servers, "acme/missing", gitTestToken))
	expectCode(t, err, "GIT-15404")
	info, err := git.Verify(ctx, tokenAccess(servers, "acme/backend", gitTestToken))
	if err != nil || info.DefaultBranch != "main" {
		t.Fatalf("Verify = %+v, %v", info, err)
	}
	expectCode(t, git.Clone(ctx, tokenAccess(servers, "acme/backend", gitTestToken), filepath.Join(t.TempDir(), "c"), entity.SyncSettings{Branch: "release"}), "GIT-03404")
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	closed := "http://" + listener.Addr().String() + "/acme/backend.git"
	_ = listener.Close()
	_, err = git.List(ctx, entity.NewGitAccess(closed, entity.GitAccount{}, ""))
	expectCode(t, err, "GIT-10503")
	_, err = git.List(ctx, entity.NewGitAccess("ssh://[bad", entity.GitAccount{}, ""))
	expectCode(t, err, "GIT-22400")
	sshAccount := entity.NewSshAccount("s", sshmode.Agent, "h", entity.SshKey{}, false, gitTestTime)
	_, err = git.List(ctx, entity.NewGitAccess(servers.HttpURL("acme/backend"), sshAccount, ""))
	expectCode(t, err, "GIT-17400")
	_, err = git.List(ctx, entity.NewGitAccess(servers.SshURL("acme/backend"), entity.GitAccount{}, ""))
	expectCode(t, err, "GIT-18400")
}

func TestGitEmptyRemoteStartsANewHistory(t *testing.T) {
	ctx := context.Background()
	servers := gitServers(t)
	if err := servers.CreateRemote("acme/empty"); err != nil {
		t.Fatal(err)
	}
	git := NewGitWebservice(failingAgent{})
	access := tokenAccess(servers, "acme/empty", "")
	info, err := git.List(ctx, access)
	if err != nil || info.DefaultBranch != "" || len(info.Branches) != 0 {
		t.Fatalf("List = %+v, %v", info, err)
	}
	clone := filepath.Join(t.TempDir(), "clone")
	if err := git.Clone(ctx, access, clone, entity.SyncSettings{Branch: "main", Subfolder: "api"}); err != nil {
		t.Fatal(err)
	}
	heads, err := git.Fetch(ctx, access, clone, "main")
	if err != nil || heads.Local != "" || heads.Remote != "" {
		t.Fatalf("Fetch = %+v, %v", heads, err)
	}
	if _, err := git.Write(ctx, clone, entity.SyncCommit{Branch: "main", Subfolder: "api", Message: "first", When: gitTestTime, Files: map[string]string{"workspace.yaml": "w\n"}}); err != nil {
		t.Fatal(err)
	}
	if err := git.Push(ctx, access, clone, "main", ""); err != nil {
		t.Fatal(err)
	}
	if count, err := servers.CommitCount("acme/empty", "main"); err != nil || count != 1 {
		t.Fatalf("commits = %d, %v", count, err)
	}
	empty, err := git.ReadVersions(ctx, clone, "missing", entity.BranchHeads{})
	if err != nil || len(empty.Base)+len(empty.Mine)+len(empty.Theirs) != 0 {
		t.Fatalf("ReadVersions = %+v, %v", empty, err)
	}
	_, err = git.ReadBlob(ctx, clone, plumbing.ZeroHash.String())
	expectCode(t, err, "GIT-57404")
	_, err = git.Fetch(ctx, access, filepath.Join(t.TempDir(), "none"), "main")
	expectCode(t, err, "GIT-42500")
	if files, err := git.ReadWorktree(ctx, filepath.Join(t.TempDir(), "none"), ""); err != nil || len(files) != 0 {
		t.Fatalf("ReadWorktree of a missing folder = %v, %v", files, err)
	}
}

func TestGitWriteRemovesAnEmptySubfolderAndKeepsSiblings(t *testing.T) {
	ctx := context.Background()
	servers := gitServers(t)
	for path, content := range map[string]string{"api/xenon/a.yaml": "a", "api/other.txt": "o"} {
		if err := servers.CommitFile("acme/backend", path, content, "seed"); err != nil {
			t.Fatal(err)
		}
	}
	git := NewGitWebservice(failingAgent{})
	access := tokenAccess(servers, "acme/backend", "")
	clone := filepath.Join(t.TempDir(), "clone")
	if err := git.Clone(ctx, access, clone, entity.SyncSettings{Branch: "main"}); err != nil {
		t.Fatal(err)
	}
	heads, _ := git.Fetch(ctx, access, clone, "main")
	if _, err := git.Write(ctx, clone, entity.SyncCommit{Branch: "main", Subfolder: "/api/xenon/", Parent: heads.Local, Message: "drop", When: gitTestTime}); err != nil {
		t.Fatal(err)
	}
	versions, err := git.ReadVersions(ctx, clone, "api", entity.BranchHeads{Local: referenceOf(t, clone)})
	if err != nil || !reflect.DeepEqual(versions.Mine, map[string]string{"other.txt": plumbing.ComputeHash(plumbing.BlobObject, []byte("o")).String()}) {
		t.Fatalf("after removing the subfolder: %+v, %v", versions.Mine, err)
	}
	if _, err := git.Write(ctx, clone, entity.SyncCommit{Branch: "main", Parent: "0123456789012345678901234567890123456789", When: gitTestTime}); err == nil {
		t.Fatal("an unknown parent must fail")
	}
}

func referenceOf(t *testing.T, clone string) string {
	t.Helper()
	heads, err := branchHeads(mustOpen(t, clone), "main")
	if err != nil {
		t.Fatal(err)
	}
	return heads.Local
}

func TestGitOverSshWithKeyFileAgentAndHostKeys(t *testing.T) {
	ctx := context.Background()
	servers := gitServers(t)
	if err := servers.CommitFile("acme/backend", "api/w.yaml", "w", "seed"); err != nil {
		t.Fatal(err)
	}
	key, err := mock.GenerateSshKey("pass")
	if err != nil {
		t.Fatal(err)
	}
	servers.AuthorizeSshKey(key.Signer.PublicKey())
	keyPath := filepath.Join(t.TempDir(), "id")
	if err := os.WriteFile(keyPath, key.PrivatePem, 0o600); err != nil {
		t.Fatal(err)
	}
	agent := mock.NewSshAgent()
	git := NewGitWebservice(agent)
	scanned, err := git.Scan(ctx, servers.SshURL("acme/backend"))
	if err != nil || scanned.Host != knownhosts.Normalize(servers.SshHost()) || scanned.Fingerprint != gossh.FingerprintSHA256(servers.SshHostKey()) || scanned.KeyType != "ssh-ed25519" {
		t.Fatalf("Scan = %+v, %v", scanned, err)
	}
	account := entity.NewSshAccount("s", sshmode.KeyFile, "127.0.0.1", entity.SshKey{Path: keyPath}, true, gitTestTime)
	trusted := entity.NewGitAccess(servers.SshURL("acme/backend"), account, "pass").WithTrustedHostKeys([]entity.HostKey{*scanned})
	info, err := git.List(ctx, trusted)
	if err != nil || info.DefaultBranch != "main" {
		t.Fatalf("List over ssh = %+v, %v", info, err)
	}
	clone := filepath.Join(t.TempDir(), "clone")
	if err := git.Clone(ctx, trusted, clone, entity.SyncSettings{Branch: "main", Subfolder: "api"}); err != nil {
		t.Fatalf("Clone over ssh: %s", describeError(err))
	}
	_, err = git.List(ctx, entity.NewGitAccess(servers.SshURL("acme/backend"), account, "pass"))
	expectCode(t, err, "GIT-26500")
	_, err = git.List(ctx, entity.NewGitAccess(servers.SshURL("acme/backend"), account, "wrong").WithTrustedHostKeys([]entity.HostKey{*scanned}))
	expectCode(t, err, "GIT-25400")
	missing := entity.NewSshAccount("s", sshmode.KeyFile, "h", entity.SshKey{Path: filepath.Join(t.TempDir(), "none")}, false, gitTestTime)
	_, err = git.List(ctx, entity.NewGitAccess(servers.SshURL("acme/backend"), missing, ""))
	expectCode(t, err, "GIT-24400")
	agentAccount := entity.NewSshAccount("s", sshmode.Agent, "127.0.0.1", entity.SshKey{}, false, gitTestTime)
	viaAgent := entity.NewGitAccess(servers.SshURL("acme/backend"), agentAccount, "").WithTrustedHostKeys([]entity.HostKey{*scanned})
	_, err = git.List(ctx, viaAgent)
	expectCode(t, err, "GIT-23400")
	other, _ := mock.GenerateSshKey("")
	if err := agent.Hold(other.Private); err != nil {
		t.Fatal(err)
	}
	_, err = git.List(ctx, viaAgent)
	expectCode(t, err, "GIT-11401")
	if err := agent.Hold(key.Private); err != nil {
		t.Fatal(err)
	}
	if _, err := git.List(ctx, viaAgent); err != nil {
		t.Fatalf("List through the agent: %v", err)
	}
	none, err := git.Scan(ctx, servers.HttpURL("acme/backend"))
	if err != nil || none != nil {
		t.Fatalf("Scan of an https URL = %v, %v", none, err)
	}
	_, err = git.Scan(ctx, "ssh://[bad")
	expectCode(t, err, "GIT-27400")
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	closed := "ssh://git@" + listener.Addr().String() + "/x.git"
	_ = listener.Close()
	_, err = git.Scan(ctx, closed)
	expectCode(t, err, "GIT-30503")
	silent, _ := net.Listen("tcp", "127.0.0.1:0")
	defer func() { _ = silent.Close() }()
	go func() {
		connection, err := silent.Accept()
		if err == nil {
			_ = connection.Close()
		}
	}()
	_, err = git.Scan(ctx, "ssh://git@"+silent.Addr().String()+"/x.git")
	expectCode(t, err, "GIT-34503")
}

func TestRemoteInfoDefaultBranch(t *testing.T) {
	main := plumbing.NewHashReference("refs/heads/main", plumbing.NewHash("1111111111111111111111111111111111111111"))
	dev := plumbing.NewHashReference("refs/heads/dev", plumbing.NewHash("2222222222222222222222222222222222222222"))
	cases := map[string]struct {
		references []*plumbing.Reference
		want       string
	}{
		"symbolic head": {[]*plumbing.Reference{plumbing.NewSymbolicReference(plumbing.HEAD, "refs/heads/dev"), main, dev}, "dev"},
		"hash head":     {[]*plumbing.Reference{plumbing.NewHashReference(plumbing.HEAD, dev.Hash()), main, dev}, "dev"},
		"detached head": {[]*plumbing.Reference{plumbing.NewHashReference(plumbing.HEAD, plumbing.ZeroHash), main}, ""},
		"no head":       {[]*plumbing.Reference{main, plumbing.NewHashReference("refs/tags/v1", main.Hash())}, ""},
	}
	for name, c := range cases {
		info := remoteInfo(c.references)
		if info.DefaultBranch != c.want {
			t.Errorf("%s: DefaultBranch = %q, want %q", name, info.DefaultBranch, c.want)
		}
	}
	if branches := remoteInfo([]*plumbing.Reference{main, dev}).Branches; !reflect.DeepEqual(branches, []string{"dev", "main"}) {
		t.Fatalf("Branches = %v", branches)
	}
}

func mustOpen(t *testing.T, clone string) *gogit.Repository {
	t.Helper()
	repository, err := gogit.PlainOpen(clone)
	if err != nil {
		t.Fatal(err)
	}
	return repository
}
