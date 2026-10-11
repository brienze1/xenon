package mock

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	"github.com/go-git/go-git/v5/storage/filesystem"
	gossh "golang.org/x/crypto/ssh"
)

const (
	gitDefaultBranch    = "main"
	gitUploadPack       = "git-upload-pack"
	gitReceivePack      = "git-receive-pack"
	gitTeammateName     = "Teammate"
	gitTeammateEmail    = "teammate@example.test"
	gitRepositorySuffix = ".git"
)

type GitHttpRequest struct {
	Repository string
	Service    string
	Username   string
	Status     int
}

type gitHttpCredentials struct {
	username    string
	token       string
	anyUsername bool
}

type gitHttpFailure struct {
	body   string
	status int
}

type GitServers struct {
	hostKey     gossh.Signer
	http        *httptest.Server
	ssh         *gliderssh.Server
	credentials *gitHttpCredentials
	failures    map[string]gitHttpFailure
	sshAddress  string
	root        string
	sshKeys     []gossh.PublicKey
	requests    []GitHttpRequest
	mu          sync.Mutex
}

var gitServersOnce sync.Once
var gitServersInstance *GitServers

func NewGitServers() *GitServers {
	gitServersOnce.Do(func() {
		servers := &GitServers{}
		_, private, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			panic(err)
		}
		servers.hostKey, err = gossh.NewSignerFromKey(private)
		if err != nil {
			panic(err)
		}
		servers.http = httptest.NewServer(http.HandlerFunc(servers.serveHttp))
		servers.startSsh()
		if err := servers.Reset(); err != nil {
			panic(err)
		}
		gitServersInstance = servers
	})
	return gitServersInstance
}

func (g *GitServers) Reset() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.root != "" {
		_ = os.RemoveAll(g.root)
	}
	root, err := os.MkdirTemp("", "xenon-git-remotes-")
	if err != nil {
		return err
	}
	g.root = root
	g.credentials = nil
	g.failures = map[string]gitHttpFailure{}
	g.sshKeys = nil
	g.requests = nil
	return nil
}

func (g *GitServers) HttpURL(name string) string {
	return g.http.URL + "/" + name + gitRepositorySuffix
}

func (g *GitServers) HttpHost() string {
	return strings.TrimPrefix(g.http.URL, "http://")
}

func (g *GitServers) SshURL(name string) string {
	return "ssh://git@" + g.sshAddress + "/" + name + gitRepositorySuffix
}

func (g *GitServers) SshHost() string {
	return g.sshAddress
}

func (g *GitServers) SshHostKey() gossh.PublicKey {
	return g.hostKey.PublicKey()
}

func (g *GitServers) RequireHttpCredentials(username, token string, anyUsername bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.credentials = &gitHttpCredentials{username: username, token: token, anyUsername: anyUsername}
}

func (g *GitServers) FailHttp(name string, status int, body string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.failures[name] = gitHttpFailure{status: status, body: body}
}

func (g *GitServers) AuthorizeSshKey(key gossh.PublicKey) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.sshKeys = append(g.sshKeys, key)
}

func (g *GitServers) HttpRequests() []GitHttpRequest {
	g.mu.Lock()
	defer g.mu.Unlock()
	return append([]GitHttpRequest{}, g.requests...)
}

func (g *GitServers) CreateRemote(name string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	_, err := g.ensureRemote(name)
	return err
}

func (g *GitServers) CommitFile(name, filePath, content, message string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, err := g.ensureRemote(name); err != nil {
		return err
	}
	worktreeFs := memfs.New()
	worktree, err := g.teammateWorktree(g.remotePath(name), worktreeFs)
	if err != nil {
		return err
	}
	if err := util.WriteFile(worktreeFs, filePath, []byte(content), 0o644); err != nil {
		return err
	}
	if _, err := worktree.Add(filePath); err != nil {
		return err
	}
	_, err = worktree.Commit(message, &git.CommitOptions{Author: &object.Signature{Name: gitTeammateName, Email: gitTeammateEmail, When: time.Now()}})
	return err
}

func (g *GitServers) ReadFile(name, branch, filePath string) (string, error) {
	commit, err := g.head(name, branch)
	if err != nil {
		return "", err
	}
	file, err := commit.File(filePath)
	if err != nil {
		return "", fmt.Errorf("the remote %s branch %s has no file %s: %w", name, branch, filePath, err)
	}
	return file.Contents()
}

func (g *GitServers) CommitCount(name, branch string) (int, error) {
	commit, err := g.head(name, branch)
	if err != nil {
		return 0, err
	}
	count := 0
	err = object.NewCommitPreorderIter(commit, nil, nil).ForEach(func(*object.Commit) error {
		count++
		return nil
	})
	return count, err
}

func (g *GitServers) LastCommit(name, branch string) (*object.Commit, error) {
	return g.head(name, branch)
}

func (g *GitServers) HistoryContains(name, text string) (bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	repository, err := git.PlainOpen(g.remotePath(name))
	if err != nil {
		return false, err
	}
	found := false
	blobs, err := repository.BlobObjects()
	if err != nil {
		return false, err
	}
	err = blobs.ForEach(func(blob *object.Blob) error {
		reader, err := blob.Reader()
		if err != nil {
			return err
		}
		defer func() { _ = reader.Close() }()
		content, err := io.ReadAll(reader)
		found = found || strings.Contains(string(content), text)
		return err
	})
	if err != nil || found {
		return found, err
	}
	commits, err := repository.CommitObjects()
	if err != nil {
		return false, err
	}
	err = commits.ForEach(func(commit *object.Commit) error {
		found = found || strings.Contains(commit.Message, text) || strings.Contains(commit.Author.String(), text)
		return nil
	})
	return found, err
}

func (g *GitServers) head(name, branch string) (*object.Commit, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	repository, err := git.PlainOpen(g.remotePath(name))
	if err != nil {
		return nil, fmt.Errorf("the remote %s does not exist: %w", name, err)
	}
	reference, err := repository.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		return nil, fmt.Errorf("the remote %s has no branch %s: %w", name, branch, err)
	}
	return repository.CommitObject(reference.Hash())
}

func (g *GitServers) teammateWorktree(location string, worktreeFs billy.Filesystem) (*git.Worktree, error) {
	storage := filesystem.NewStorage(osfs.New(location), cache.NewObjectLRUDefault())
	opened, err := git.Open(storage, worktreeFs)
	if err != nil {
		return nil, err
	}
	worktree, err := opened.Worktree()
	if err != nil {
		return nil, err
	}
	branch := plumbing.NewBranchReferenceName(gitDefaultBranch)
	if _, err := opened.Reference(branch, true); err == nil {
		if err := worktree.Checkout(&git.CheckoutOptions{Branch: branch, Force: true}); err != nil {
			return nil, err
		}
	}
	return worktree, nil
}

func (g *GitServers) ensureRemote(name string) (*git.Repository, error) {
	location := g.remotePath(name)
	if repository, err := git.PlainOpen(location); err == nil {
		return repository, nil
	}
	return git.PlainInitWithOptions(location, &git.PlainInitOptions{
		Bare:        true,
		InitOptions: git.InitOptions{DefaultBranch: plumbing.NewBranchReferenceName(gitDefaultBranch)},
	})
}

func (g *GitServers) remotePath(name string) string {
	return filepath.Join(g.root, filepath.FromSlash(name)+gitRepositorySuffix)
}

func (g *GitServers) transport() transport.Transport {
	return server.NewServer(g)
}

func (g *GitServers) Load(endpoint *transport.Endpoint) (storer.Storer, error) {
	g.mu.Lock()
	location := g.remotePath(strings.TrimSuffix(strings.TrimPrefix(endpoint.Path, "/"), gitRepositorySuffix))
	g.mu.Unlock()
	if _, err := os.Stat(filepath.Join(location, "config")); err != nil {
		return nil, transport.ErrRepositoryNotFound
	}
	return filesystem.NewStorage(osfs.New(location), cache.NewObjectLRUDefault()), nil
}

func (g *GitServers) serveHttp(w http.ResponseWriter, r *http.Request) {
	repository, service := splitGitHttpPath(r)
	username, token, _ := r.BasicAuth()
	status := g.authorizeHttp(repository, username, token)
	g.record(GitHttpRequest{Repository: repository, Service: service, Username: username, Status: status})
	if status != http.StatusOK {
		g.refuseHttp(w, repository, status)
		return
	}
	endpoint, err := transport.NewEndpoint("/" + repository + gitRepositorySuffix)
	if err != nil || service == "" {
		http.Error(w, "not a git endpoint", http.StatusNotFound)
		return
	}
	switch {
	case r.Method == http.MethodGet:
		g.advertise(w, endpoint, service)
	case service == gitUploadPack:
		g.uploadPack(w, r, endpoint)
	default:
		g.receivePack(w, r, endpoint)
	}
}

func (g *GitServers) authorizeHttp(repository, username, token string) int {
	g.mu.Lock()
	defer g.mu.Unlock()
	if failure, failing := g.failures[repository]; failing {
		return failure.status
	}
	expected := g.credentials
	if expected == nil {
		return http.StatusOK
	}
	if token != expected.token || (!expected.anyUsername && username != expected.username) || (expected.anyUsername && username == "") {
		return http.StatusUnauthorized
	}
	return http.StatusOK
}

func (g *GitServers) refuseHttp(w http.ResponseWriter, repository string, status int) {
	g.mu.Lock()
	failure, failing := g.failures[repository]
	g.mu.Unlock()
	if status == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", `Basic realm="xenon-test-git"`)
	}
	body := "Authentication failed"
	if failing {
		body = failure.body
	}
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func (g *GitServers) record(request GitHttpRequest) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.requests = append(g.requests, request)
}

func (g *GitServers) advertise(w http.ResponseWriter, endpoint *transport.Endpoint, service string) {
	references, err := g.advertisedReferences(endpoint, service)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	references.Prefix = [][]byte{[]byte("# service=" + service), pktline.Flush}
	w.Header().Set("Content-Type", "application/x-"+service+"-advertisement")
	w.Header().Set("Cache-Control", "no-cache")
	_ = references.Encode(w)
}

func (g *GitServers) advertisedReferences(endpoint *transport.Endpoint, service string) (*packp.AdvRefs, error) {
	if service == gitUploadPack {
		session, err := g.transport().NewUploadPackSession(endpoint, nil)
		if err != nil {
			return nil, err
		}
		return session.AdvertisedReferences()
	}
	session, err := g.transport().NewReceivePackSession(endpoint, nil)
	if err != nil {
		return nil, err
	}
	return session.AdvertisedReferences()
}

func (g *GitServers) uploadPack(w http.ResponseWriter, r *http.Request, endpoint *transport.Endpoint) {
	session, err := g.transport().NewUploadPackSession(endpoint, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	request := packp.NewUploadPackRequest()
	if err := request.Decode(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response, err := session.UploadPack(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = response.Close() }()
	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	_ = response.Encode(w)
}

func (g *GitServers) receivePack(w http.ResponseWriter, r *http.Request, endpoint *transport.Endpoint) {
	session, err := g.transport().NewReceivePackSession(endpoint, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if _, err := session.AdvertisedReferences(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	request := packp.NewReferenceUpdateRequest()
	if err := request.Decode(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	g.mu.Lock()
	status, _ := session.ReceivePack(r.Context(), request)
	g.mu.Unlock()
	w.Header().Set("Content-Type", "application/x-git-receive-pack-result")
	if status != nil {
		_ = status.Encode(w)
	}
}

func (g *GitServers) startSsh() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	g.sshAddress = listener.Addr().String()
	g.ssh = &gliderssh.Server{
		Handler:          g.serveSsh,
		PublicKeyHandler: g.authorizeSsh,
	}
	g.ssh.AddHostKey(g.hostKey)
	go func() { _ = g.ssh.Serve(listener) }()
}

func (g *GitServers) authorizeSsh(_ gliderssh.Context, key gliderssh.PublicKey) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, authorized := range g.sshKeys {
		if gliderssh.KeysEqual(authorized, key) {
			return true
		}
	}
	return false
}

func (g *GitServers) serveSsh(session gliderssh.Session) {
	command := session.Command()
	if len(command) != 2 {
		_ = session.Exit(1)
		return
	}
	endpoint, err := transport.NewEndpoint("/" + strings.TrimPrefix(path.Clean(command[1]), "/"))
	if err != nil {
		_ = session.Exit(1)
		return
	}
	if err := g.serveSshCommand(session.Context(), session, command[0], endpoint); err != nil {
		_, _ = io.WriteString(session.Stderr(), err.Error())
		_ = session.Exit(1)
		return
	}
	_ = session.Exit(0)
}

func (g *GitServers) serveSshCommand(ctx context.Context, stream io.ReadWriter, command string, endpoint *transport.Endpoint) error {
	switch command {
	case gitUploadPack:
		session, err := g.transport().NewUploadPackSession(endpoint, nil)
		if err != nil {
			return err
		}
		references, err := session.AdvertisedReferences()
		if err != nil {
			return err
		}
		if err := references.Encode(stream); err != nil {
			return err
		}
		request := packp.NewUploadPackRequest()
		if err := request.Decode(stream); err != nil {
			return nil
		}
		response, err := session.UploadPack(ctx, request)
		if err != nil {
			return err
		}
		defer func() { _ = response.Close() }()
		return response.Encode(stream)
	case gitReceivePack:
		session, err := g.transport().NewReceivePackSession(endpoint, nil)
		if err != nil {
			return err
		}
		references, err := session.AdvertisedReferences()
		if err != nil {
			return err
		}
		if err := references.Encode(stream); err != nil {
			return err
		}
		request := packp.NewReferenceUpdateRequest()
		if err := request.Decode(struct{ io.Reader }{stream}); err != nil {
			return nil
		}
		g.mu.Lock()
		status, _ := session.ReceivePack(ctx, request)
		g.mu.Unlock()
		if status != nil {
			return status.Encode(stream)
		}
		return nil
	default:
		return errors.New("unsupported command " + command)
	}
}

func splitGitHttpPath(r *http.Request) (string, string) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/")
	for _, suffix := range []string{"/info/refs", "/" + gitUploadPack, "/" + gitReceivePack} {
		if strings.HasSuffix(trimmed, suffix) {
			repository := strings.TrimSuffix(strings.TrimSuffix(trimmed, suffix), gitRepositorySuffix)
			service := strings.TrimPrefix(suffix, "/")
			if suffix == "/info/refs" {
				service = r.URL.Query().Get("service")
			}
			return repository, service
		}
	}
	return strings.TrimSuffix(trimmed, gitRepositorySuffix), ""
}
