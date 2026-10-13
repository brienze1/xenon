package webservice

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
	"golang.org/x/crypto/ssh"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const hostKeyScanTimeout = 15 * time.Second

var errHostKeyCaptured = errors.New("host key captured")

func (g *gitWebservice) List(ctx context.Context, access entity.GitAccess) (*entity.RemoteInfo, error) {
	logger.Debug(ctx, "Started", access)

	info, err := g.list(ctx, access, syncRemoteCodes)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", info)
	return info, nil
}

func (g *gitWebservice) Verify(ctx context.Context, access entity.GitAccess) (*entity.RemoteInfo, error) {
	logger.Debug(ctx, "Started", access)

	info, err := g.list(ctx, access, remoteCodes{unauthorized: "ACC-02401", restricted: "ACC-61403", notFound: "ACC-07404", offline: "ACC-08503", failed: "ACC-09500"})
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", info)
	return info, nil
}

func (g *gitWebservice) Scan(ctx context.Context, repository string) (*entity.HostKey, error) {
	logger.Debug(ctx, "Started", map[string]any{"repository": repository})

	endpoint, err := transport.NewEndpoint(repository)
	if err != nil {
		return nil, errs.BadRequestError("the repository URL is not valid: "+repository, "GIT-27400")
	}
	if endpoint.Protocol != sshProtocol {
		logger.Debug(ctx, "Finished", map[string]any{"repository": repository, "ssh": false})
		return nil, nil
	}
	address, host := sshAddress(endpoint)
	var captured ssh.PublicKey
	dialer := net.Dialer{Timeout: hostKeyScanTimeout}
	connection, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, remoteFailure("the SSH host could not be reached: "+err.Error(), "GIT-30503", http.StatusServiceUnavailable, offlineNextStep)
	}
	defer func() { _ = connection.Close() }()
	_ = connection.SetDeadline(time.Now().Add(hostKeyScanTimeout))
	_, _, _, err = ssh.NewClientConn(connection, address, &ssh.ClientConfig{
		User: defaultSshUser,
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			captured = key
			return errHostKeyCaptured
		},
	})
	if captured == nil {
		return nil, remoteFailure("the SSH host did not present a host key: "+err.Error(), "GIT-34503", http.StatusServiceUnavailable, offlineNextStep)
	}
	key := newHostKey(host, captured)

	logger.Debug(ctx, "Finished", key)
	return &key, nil
}

func (g *gitWebservice) Clone(ctx context.Context, access entity.GitAccess, clonePath string, settings entity.SyncSettings) error {
	logger.Debug(ctx, "Started", access, settings, map[string]any{"clonePath": clonePath})

	defer g.lock(clonePath)()
	auth, closer, err := g.auth(access)
	if err != nil {
		return err
	}
	defer func() { _ = closer.Close() }()
	branch := plumbing.NewBranchReferenceName(settings.Branch)
	_, err = git.PlainCloneContext(ctx, clonePath, false, &git.CloneOptions{
		URL: access.Repository, Auth: auth, RemoteName: originRemote, ReferenceName: branch,
		SingleBranch: true, NoCheckout: true, Tags: git.NoTags,
	})
	switch {
	case errors.Is(err, transport.ErrEmptyRemoteRepository):
		_ = os.RemoveAll(clonePath)
		err = initEmptyClone(clonePath, access.Repository, branch)
		if err != nil {
			return errs.FileError(err, "GIT-36500")
		}
	case err != nil && isMissingReference(err):
		return remoteFailure("the branch "+settings.Branch+" does not exist on the remote", "GIT-03404", http.StatusNotFound, "Pick another branch")
	case err != nil:
		return remoteError(err, syncRemoteCodes)
	default:
		if err := checkout(clonePath, settings.Branch, settings.Subfolder); err != nil {
			return errs.InternalServerError(err, "GIT-41500", "the clone could not be checked out")
		}
	}

	logger.Debug(ctx, "Finished", settings, map[string]any{"clonePath": clonePath})
	return nil
}

func (g *gitWebservice) Fetch(ctx context.Context, access entity.GitAccess, clonePath, branch string) (*entity.BranchHeads, error) {
	logger.Debug(ctx, "Started", access, map[string]any{"clonePath": clonePath, "branch": branch})

	defer g.lock(clonePath)()
	repository, err := git.PlainOpen(clonePath)
	if err != nil {
		return nil, errs.FileError(err, "GIT-42500")
	}
	auth, closer, err := g.auth(access)
	if err != nil {
		return nil, err
	}
	defer func() { _ = closer.Close() }()
	refSpec := config.RefSpec("+refs/heads/" + branch + ":refs/remotes/" + originRemote + "/" + branch)
	err = repository.FetchContext(ctx, &git.FetchOptions{RemoteName: originRemote, RefSpecs: []config.RefSpec{refSpec}, Auth: auth, Tags: git.NoTags})
	switch {
	case err == nil, errors.Is(err, git.NoErrAlreadyUpToDate), errors.Is(err, transport.ErrEmptyRemoteRepository):
	case isMissingReference(err):
		return nil, remoteFailure("the branch "+branch+" does not exist on the remote", "GIT-20404", http.StatusNotFound, "Pick another branch")
	default:
		return nil, remoteError(err, syncRemoteCodes)
	}
	heads, err := branchHeads(repository, branch)
	if err != nil {
		return nil, errs.InternalServerError(err, "GIT-47500", "the merge base could not be computed")
	}

	logger.Debug(ctx, "Finished", heads)
	return heads, nil
}

func (g *gitWebservice) Push(ctx context.Context, access entity.GitAccess, clonePath, branch, expectedRemote string) error {
	logger.Debug(ctx, "Started", access, map[string]any{"clonePath": clonePath, "branch": branch, "expectedRemote": expectedRemote})

	defer g.lock(clonePath)()
	repository, err := git.PlainOpen(clonePath)
	if err != nil {
		return errs.FileError(err, "GIT-48500")
	}
	auth, closer, err := g.auth(access)
	if err != nil {
		return err
	}
	defer func() { _ = closer.Close() }()
	ref := plumbing.NewBranchReferenceName(branch)
	options := &git.PushOptions{RemoteName: originRemote, RefSpecs: []config.RefSpec{config.RefSpec(ref + ":" + ref)}, Auth: auth}
	if expectedRemote != "" {
		options.ForceWithLease = &git.ForceWithLease{RefName: ref, Hash: plumbing.NewHash(expectedRemote)}
	}
	err = repository.PushContext(ctx, options)
	switch {
	case err == nil, errors.Is(err, git.NoErrAlreadyUpToDate):
	default:
		return pushError(err)
	}

	logger.Debug(ctx, "Finished", map[string]any{"clonePath": clonePath, "branch": branch})
	return nil
}

func (g *gitWebservice) list(ctx context.Context, access entity.GitAccess, codes remoteCodes) (*entity.RemoteInfo, error) {
	auth, closer, err := g.auth(access)
	if err != nil {
		return nil, err
	}
	defer func() { _ = closer.Close() }()
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{Name: originRemote, URLs: []string{access.Repository}})
	references, err := remote.ListContext(ctx, &git.ListOptions{Auth: auth})
	if errors.Is(err, transport.ErrEmptyRemoteRepository) {
		return &entity.RemoteInfo{Branches: []string{}}, nil
	}
	if err != nil {
		return nil, remoteError(err, codes)
	}
	return remoteInfo(references), nil
}

func remoteInfo(references []*plumbing.Reference) *entity.RemoteInfo {
	info := &entity.RemoteInfo{Branches: []string{}}
	var head *plumbing.Reference
	hashes := map[string]plumbing.Hash{}
	for _, reference := range references {
		switch {
		case reference.Name() == plumbing.HEAD:
			head = reference
		case reference.Name().IsBranch():
			info.Branches = append(info.Branches, reference.Name().Short())
			hashes[reference.Name().Short()] = reference.Hash()
		}
	}
	sort.Strings(info.Branches)
	if head == nil {
		return info
	}
	if head.Type() == plumbing.SymbolicReference {
		info.DefaultBranch = head.Target().Short()
		return info
	}
	for _, branch := range info.Branches {
		if hashes[branch] == head.Hash() {
			info.DefaultBranch = branch
			return info
		}
	}
	return info
}

func initEmptyClone(clonePath, repository string, branch plumbing.ReferenceName) error {
	created, err := git.PlainInitWithOptions(clonePath, &git.PlainInitOptions{InitOptions: git.InitOptions{DefaultBranch: branch}})
	if err != nil {
		return err
	}
	_, err = created.CreateRemote(&config.RemoteConfig{
		Name:  originRemote,
		URLs:  []string{repository},
		Fetch: []config.RefSpec{config.RefSpec("+refs/heads/" + branch.Short() + ":refs/remotes/" + originRemote + "/" + branch.Short())},
	})
	return err
}

func branchHeads(repository *git.Repository, branch string) (*entity.BranchHeads, error) {
	heads := &entity.BranchHeads{
		Local:  referenceHash(repository, plumbing.NewBranchReferenceName(branch)),
		Remote: referenceHash(repository, plumbing.NewRemoteReferenceName(originRemote, branch)),
	}
	local, err := optionalCommit(repository, heads.Local)
	if err != nil {
		return nil, err
	}
	remote, err := optionalCommit(repository, heads.Remote)
	if err != nil {
		return nil, err
	}
	if local != nil && remote != nil {
		bases, err := local.MergeBase(remote)
		if err != nil {
			return nil, err
		}
		if len(bases) > 0 {
			heads.Base = bases[0].Hash.String()
		}
	}
	if heads.Ahead, err = commitsSince(local, heads.Base); err != nil {
		return nil, err
	}
	if heads.Behind, err = commitsSince(remote, heads.Base); err != nil {
		return nil, err
	}
	return heads, nil
}

func optionalCommit(repository *git.Repository, hash string) (*object.Commit, error) {
	if hash == "" {
		return nil, nil
	}
	return repository.CommitObject(plumbing.NewHash(hash))
}

func commitsSince(head *object.Commit, base string) (int, error) {
	if head == nil {
		return 0, nil
	}
	ignore := []plumbing.Hash{}
	if base != "" {
		ignore = append(ignore, plumbing.NewHash(base))
	}
	count := 0
	err := object.NewCommitPreorderIter(head, nil, ignore).ForEach(func(*object.Commit) error {
		count++
		return nil
	})
	return count, err
}

func referenceHash(repository *git.Repository, name plumbing.ReferenceName) string {
	reference, err := repository.Reference(name, true)
	if err != nil {
		return ""
	}
	return reference.Hash().String()
}

func isMissingReference(err error) bool {
	return errors.Is(err, plumbing.ErrReferenceNotFound) || strings.Contains(err.Error(), "couldn't find remote ref")
}
