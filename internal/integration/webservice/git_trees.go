package webservice

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const gitMetadataDir = ".git"

type treeDraft struct {
	files map[string]plumbing.Hash
	dirs  map[string]*treeDraft
}

func (g *gitWebservice) ReadVersions(ctx context.Context, clonePath, subfolder string, heads entity.BranchHeads) (*entity.TreeVersions, error) {
	logger.Debug(ctx, "Started", heads, map[string]any{"clonePath": clonePath, "subfolder": subfolder})

	defer g.lock(clonePath)()
	repository, err := git.PlainOpen(clonePath)
	if err != nil {
		return nil, errs.FileError(err, "GIT-53500")
	}
	versions := &entity.TreeVersions{}
	for _, version := range []struct {
		target *map[string]string
		hash   string
	}{{&versions.Base, heads.Base}, {&versions.Mine, heads.Local}, {&versions.Theirs, heads.Remote}} {
		files, err := commitFiles(repository, version.hash, subfolder)
		if err != nil {
			return nil, errs.InternalServerError(err, "GIT-54500", "the commit "+version.hash+" could not be read")
		}
		*version.target = files
	}

	logger.Debug(ctx, "Finished", map[string]any{"base": len(versions.Base), "mine": len(versions.Mine), "theirs": len(versions.Theirs)})
	return versions, nil
}

func (g *gitWebservice) ReadWorktree(ctx context.Context, clonePath, subfolder string) (map[string]string, error) {
	logger.Debug(ctx, "Started", map[string]any{"clonePath": clonePath, "subfolder": subfolder})

	root := filepath.Join(clonePath, filepath.FromSlash(subfolder))
	files := map[string]string{}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case entry.IsDir() && entry.Name() == gitMetadataDir:
			return filepath.SkipDir
		case !entry.Type().IsRegular():
			return nil
		}
		content, err := os.ReadFile(current)
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(root, current)
		files[filepath.ToSlash(relative)] = plumbing.ComputeHash(plumbing.BlobObject, content).String()
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, errs.FileError(err, "GIT-55500")
	}

	logger.Debug(ctx, "Finished", map[string]any{"files": len(files)})
	return files, nil
}

func (g *gitWebservice) ReadBlob(ctx context.Context, clonePath, hash string) ([]byte, error) {
	logger.Debug(ctx, "Started", map[string]any{"clonePath": clonePath, "hash": hash})

	defer g.lock(clonePath)()
	repository, err := git.PlainOpen(clonePath)
	if err != nil {
		return nil, errs.FileError(err, "GIT-56500")
	}
	blob, err := repository.BlobObject(plumbing.NewHash(hash))
	if err != nil {
		return nil, errs.NotFoundError("the file version "+hash+" is not in the clone", "GIT-57404")
	}
	reader, err := blob.Reader()
	if err != nil {
		return nil, errs.InternalServerError(err, "GIT-58500", "the file version could not be read")
	}
	defer func() { _ = reader.Close() }()
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, errs.InternalServerError(err, "GIT-59500", "the file version could not be read")
	}

	logger.Debug(ctx, "Finished", map[string]any{"hash": hash, "bytes": len(content)})
	return content, nil
}

func (g *gitWebservice) Write(ctx context.Context, clonePath string, commit entity.SyncCommit) (string, error) {
	logger.Debug(ctx, "Started", commit, map[string]any{"clonePath": clonePath})

	defer g.lock(clonePath)()
	repository, err := git.PlainOpen(clonePath)
	if err != nil {
		return "", errs.FileError(err, "GIT-60500")
	}
	hash, err := writeCommit(repository.Storer, commit)
	if err != nil {
		return "", errs.InternalServerError(err, "GIT-61500", "the sync commit could not be written")
	}

	logger.Debug(ctx, "Finished", map[string]any{"commit": hash})
	return hash, nil
}

func (g *gitWebservice) Checkout(ctx context.Context, clonePath, branch, subfolder string) error {
	logger.Debug(ctx, "Started", map[string]any{"clonePath": clonePath, "branch": branch, "subfolder": subfolder})

	defer g.lock(clonePath)()
	if err := checkout(clonePath, branch, subfolder); err != nil {
		return errs.InternalServerError(err, "GIT-62500", "the clone could not be checked out")
	}

	logger.Debug(ctx, "Finished", map[string]any{"clonePath": clonePath, "branch": branch})
	return nil
}

func checkout(clonePath, branch, subfolder string) error {
	repository, err := git.PlainOpen(clonePath)
	if err != nil {
		return err
	}
	worktree, err := repository.Worktree()
	if err != nil {
		return err
	}
	options := &git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(branch), Force: true}
	if subfolder != "" {
		options.SparseCheckoutDirectories = []string{subfolder}
	}
	return worktree.Checkout(options)
}

func commitFiles(repository *git.Repository, hash, subfolder string) (map[string]string, error) {
	files := map[string]string{}
	if hash == "" {
		return files, nil
	}
	commit, err := repository.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	if subfolder != "" {
		if tree, err = tree.Tree(subfolder); err != nil {
			return files, nil
		}
	}
	err = tree.Files().ForEach(func(file *object.File) error {
		files[file.Name] = file.Hash.String()
		return nil
	})
	return files, err
}

func writeCommit(store storer.EncodedObjectStorer, commit entity.SyncCommit) (string, error) {
	draft := &treeDraft{files: map[string]plumbing.Hash{}, dirs: map[string]*treeDraft{}}
	for name, content := range commit.Files {
		blob, err := writeBlob(store, content)
		if err != nil {
			return "", err
		}
		draft.add(strings.Split(name, "/"), blob)
	}
	subtree, err := draft.write(store)
	if err != nil {
		return "", err
	}
	var parentTree *object.Tree
	parents := []plumbing.Hash{}
	if commit.Parent != "" {
		parent, err := object.GetCommit(store, plumbing.NewHash(commit.Parent))
		if err != nil {
			return "", err
		}
		if parentTree, err = parent.Tree(); err != nil {
			return "", err
		}
		parents = append(parents, parent.Hash)
	}
	root, err := replaceSubtree(store, parentTree, splitSubfolder(commit.Subfolder), subtree, len(commit.Files) == 0)
	if err != nil {
		return "", err
	}
	signature := object.Signature{Name: commit.AuthorName, Email: commit.AuthorEmail, When: commit.When}
	created := &object.Commit{Author: signature, Committer: signature, Message: commit.Message, TreeHash: root, ParentHashes: parents}
	hash, err := storeObject(store, created)
	if err != nil {
		return "", err
	}
	if references, ok := store.(storer.ReferenceStorer); ok {
		if err := references.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName(commit.Branch), hash)); err != nil {
			return "", err
		}
	}
	return hash.String(), nil
}

func (t *treeDraft) add(segments []string, blob plumbing.Hash) {
	if len(segments) == 1 {
		t.files[segments[0]] = blob
		return
	}
	child, found := t.dirs[segments[0]]
	if !found {
		child = &treeDraft{files: map[string]plumbing.Hash{}, dirs: map[string]*treeDraft{}}
		t.dirs[segments[0]] = child
	}
	child.add(segments[1:], blob)
}

func (t *treeDraft) write(store storer.EncodedObjectStorer) (plumbing.Hash, error) {
	entries := make([]object.TreeEntry, 0)
	for name, blob := range t.files {
		entries = append(entries, object.TreeEntry{Name: name, Mode: filemode.Regular, Hash: blob})
	}
	for name, child := range t.dirs {
		hash, err := child.write(store)
		if err != nil {
			return plumbing.ZeroHash, err
		}
		entries = append(entries, object.TreeEntry{Name: name, Mode: filemode.Dir, Hash: hash})
	}
	return writeTree(store, entries)
}

func replaceSubtree(store storer.EncodedObjectStorer, tree *object.Tree, segments []string, subtree plumbing.Hash, empty bool) (plumbing.Hash, error) {
	if len(segments) == 0 {
		return subtree, nil
	}
	entries := make([]object.TreeEntry, 0)
	var current *object.Tree
	if tree != nil {
		for _, entry := range tree.Entries {
			if entry.Name != segments[0] {
				entries = append(entries, entry)
				continue
			}
			if entry.Mode == filemode.Dir {
				current, _ = object.GetTree(store, entry.Hash)
			}
		}
	}
	child, err := replaceSubtree(store, current, segments[1:], subtree, empty)
	if err != nil {
		return plumbing.ZeroHash, err
	}
	childTree, err := object.GetTree(store, child)
	if err != nil {
		return plumbing.ZeroHash, err
	}
	if !(empty && len(childTree.Entries) == 0) {
		entries = append(entries, object.TreeEntry{Name: segments[0], Mode: filemode.Dir, Hash: child})
	}
	return writeTree(store, entries)
}

func writeTree(store storer.EncodedObjectStorer, entries []object.TreeEntry) (plumbing.Hash, error) {
	sort.Slice(entries, func(i, j int) bool { return treeSortKey(entries[i]) < treeSortKey(entries[j]) })
	return storeObject(store, &object.Tree{Entries: entries})
}

func treeSortKey(entry object.TreeEntry) string {
	if entry.Mode == filemode.Dir {
		return entry.Name + "/"
	}
	return entry.Name
}

func writeBlob(store storer.EncodedObjectStorer, content string) (plumbing.Hash, error) {
	encoded := store.NewEncodedObject()
	encoded.SetType(plumbing.BlobObject)
	writer, err := encoded.Writer()
	if err != nil {
		return plumbing.ZeroHash, err
	}
	if _, err := io.WriteString(writer, content); err != nil {
		return plumbing.ZeroHash, err
	}
	if err := writer.Close(); err != nil {
		return plumbing.ZeroHash, err
	}
	return store.SetEncodedObject(encoded)
}

type encodable interface {
	Encode(plumbing.EncodedObject) error
}

func storeObject(store storer.EncodedObjectStorer, value encodable) (plumbing.Hash, error) {
	encoded := store.NewEncodedObject()
	if err := value.Encode(encoded); err != nil {
		return plumbing.ZeroHash, err
	}
	return store.SetEncodedObject(encoded)
}

func splitSubfolder(subfolder string) []string {
	cleaned := strings.Trim(path.Clean("/"+subfolder), "/")
	if cleaned == "" {
		return nil
	}
	return strings.Split(cleaned, "/")
}
