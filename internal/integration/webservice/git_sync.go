package webservice

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

func (g *gitWebservice) Claim(ctx context.Context, localId string) (func(), error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	value, _ := g.operations.LoadOrStore(localId, &sync.Mutex{})
	mutex := value.(*sync.Mutex)
	if !mutex.TryLock() {
		return nil, errs.ConflictError(busyDescription, "GIT-05409")
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId})
	return mutex.Unlock, nil
}

func (g *gitWebservice) ReadHeads(ctx context.Context, clonePath, branch string) (*entity.BranchHeads, error) {
	logger.Debug(ctx, "Started", map[string]any{"clonePath": clonePath, "branch": branch})

	defer g.lock(clonePath)()
	repository, err := git.PlainOpen(clonePath)
	if err != nil {
		return nil, errs.FileError(err, "GIT-63500")
	}
	heads, err := branchHeads(repository, branch)
	if err != nil {
		return nil, errs.InternalServerError(err, "GIT-64500", "the branch heads could not be read")
	}

	logger.Debug(ctx, "Finished", heads)
	return heads, nil
}

func (g *gitWebservice) MoveBranch(ctx context.Context, clonePath, branch, hash string) error {
	logger.Debug(ctx, "Started", map[string]any{"clonePath": clonePath, "branch": branch, "hash": hash})

	defer g.lock(clonePath)()
	repository, err := git.PlainOpen(clonePath)
	if err != nil {
		return errs.FileError(err, "GIT-65500")
	}
	reference := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branch), plumbing.NewHash(hash))
	if err := repository.Storer.SetReference(reference); err != nil {
		return errs.InternalServerError(err, "GIT-66500", "the branch could not be moved")
	}

	logger.Debug(ctx, "Finished", map[string]any{"clonePath": clonePath, "branch": branch, "hash": hash})
	return nil
}

func (g *gitWebservice) ListMessages(ctx context.Context, clonePath, base, head string) ([]string, error) {
	logger.Debug(ctx, "Started", map[string]any{"clonePath": clonePath, "base": base, "head": head})

	defer g.lock(clonePath)()
	repository, err := git.PlainOpen(clonePath)
	if err != nil {
		return nil, errs.FileError(err, "GIT-67500")
	}
	commit, err := repository.CommitObject(plumbing.NewHash(head))
	if err != nil {
		return nil, errs.InternalServerError(err, "GIT-68500", "the local commits could not be read")
	}
	ignore := []plumbing.Hash{}
	if base != "" {
		ignore = append(ignore, plumbing.NewHash(base))
	}
	messages := make([]string, 0)
	err = object.NewCommitPreorderIter(commit, nil, ignore).ForEach(func(current *object.Commit) error {
		messages = append([]string{current.Message}, messages...)
		return nil
	})
	if err != nil {
		return nil, errs.InternalServerError(err, "GIT-69500", "the local commits could not be read")
	}

	logger.Debug(ctx, "Finished", map[string]any{"messages": len(messages)})
	return messages, nil
}

func (g *gitWebservice) ReadWorktreeFiles(ctx context.Context, clonePath, subfolder string, paths []string) (map[string][]byte, error) {
	logger.Debug(ctx, "Started", map[string]any{"clonePath": clonePath, "subfolder": subfolder, "paths": len(paths)})

	files := make(map[string][]byte, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(worktreePath(clonePath, subfolder, path))
		if err != nil {
			return nil, errs.FileError(err, "GIT-70500")
		}
		files[path] = content
	}

	logger.Debug(ctx, "Finished", map[string]any{"files": len(files)})
	return files, nil
}

func (g *gitWebservice) StoreWorktreeBlobs(ctx context.Context, clonePath, subfolder string, paths []string) error {
	logger.Debug(ctx, "Started", map[string]any{"clonePath": clonePath, "subfolder": subfolder, "paths": len(paths)})

	defer g.lock(clonePath)()
	repository, err := git.PlainOpen(clonePath)
	if err != nil {
		return errs.FileError(err, "GIT-71500")
	}
	for _, path := range paths {
		content, err := os.ReadFile(worktreePath(clonePath, subfolder, path))
		if err != nil {
			return errs.FileError(err, "GIT-72500")
		}
		if _, err := writeBlob(repository.Storer, string(content)); err != nil {
			return errs.InternalServerError(err, "GIT-73500", "the file "+path+" could not be stored in the clone")
		}
	}

	logger.Debug(ctx, "Finished", map[string]any{"paths": len(paths)})
	return nil
}

func (g *gitWebservice) StoreBlob(ctx context.Context, clonePath string, content []byte) (string, error) {
	logger.Debug(ctx, "Started", map[string]any{"clonePath": clonePath, "bytes": len(content)})

	defer g.lock(clonePath)()
	repository, err := git.PlainOpen(clonePath)
	if err != nil {
		return "", errs.FileError(err, "GIT-74500")
	}
	hash, err := writeBlob(repository.Storer, string(content))
	if err != nil {
		return "", errs.InternalServerError(err, "GIT-75500", "the merged file could not be stored in the clone")
	}

	logger.Debug(ctx, "Finished", map[string]any{"hash": hash.String()})
	return hash.String(), nil
}

func worktreePath(clonePath, subfolder, path string) string {
	return filepath.Join(clonePath, filepath.FromSlash(subfolder), filepath.FromSlash(path))
}
