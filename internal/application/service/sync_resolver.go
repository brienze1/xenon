package service

import (
	"context"
	"path"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
)

type resolvedPull struct {
	tree   map[string]string
	remote map[string]string
	update entity.WorktreeUpdate
}

type syncResolver struct {
	readTreeVersions    usecase.ReadTreeVersions
	storeBlobs          usecase.StoreBlobs
	convertSyncDocument usecase.ConvertSyncDocument
	changes             syncChanges
	merger              syncMerger
}

func (s syncResolver) resolve(ctx context.Context, linked entity.LinkedWorkspace, session entity.SyncSession, snapshot syncSnapshot) (*resolvedPull, error) {
	remote, err := s.changes.tree(ctx, linked, session.RemoteHead)
	if err != nil {
		return nil, err
	}
	merged := remote
	if session.Replay {
		target, err := s.merger.merge(ctx, linked, session.Heads(), session.FieldResolutions())
		if err != nil {
			return nil, err
		}
		merged = target.tree
	}
	tree := session.ResolvedTree(merged)
	for _, conflict := range session.CommittedConflicts() {
		if conflict.KeepsBoth() {
			if err := s.addCommittedCopy(ctx, linked, conflict, tree); err != nil {
				return nil, err
			}
		}
	}
	update := entity.NewWorktreeUpdate(snapshot.head, tree, snapshot.worktree)
	for _, conflict := range session.UncommittedConflicts() {
		update = update.Resolved(conflict.Path, conflict.Result(), snapshot.worktree)
		if !conflict.KeepsBoth() {
			continue
		}
		copyPath := usecase.TheirCopyPath(conflict, tree, snapshot.worktree)
		hash, err := s.theirCopy(ctx, linked, conflict)
		if err != nil {
			return nil, err
		}
		update = update.Resolved(copyPath, hash, snapshot.worktree)
	}
	return &resolvedPull{tree: tree, remote: remote, update: update}, nil
}

func (s syncResolver) addCommittedCopy(ctx context.Context, linked entity.LinkedWorkspace, conflict entity.ConflictItem, tree map[string]string) error {
	copyPath := usecase.TheirCopyPath(conflict, tree)
	hash, err := s.theirCopy(ctx, linked, conflict)
	if err != nil {
		return err
	}
	tree[copyPath] = hash
	for _, parent := range usecase.OrderParents(conflict.Path) {
		if _, found := tree[parent]; !found {
			continue
		}
		listed, err := s.listAfter(ctx, linked, tree[parent], path.Base(conflict.Path), path.Base(copyPath))
		if err != nil {
			return err
		}
		tree[parent] = listed
	}
	return nil
}

func (s syncResolver) theirCopy(ctx context.Context, linked entity.LinkedWorkspace, conflict entity.ConflictItem) (string, error) {
	document, err := s.document(ctx, linked, conflict.Theirs)
	if err != nil {
		return "", err
	}
	return s.store(ctx, linked, document.Renamed(idutils.NewUlid(), conflict.CopyName()))
}

func (s syncResolver) listAfter(ctx context.Context, linked entity.LinkedWorkspace, parentHash, existing, added string) (string, error) {
	document, err := s.document(ctx, linked, parentHash)
	if err != nil {
		return "", err
	}
	listed, changed := document.WithOrderEntryAfter(existing, added)
	if !changed {
		return parentHash, nil
	}
	return s.store(ctx, linked, listed)
}

func (s syncResolver) document(ctx context.Context, linked entity.LinkedWorkspace, hash string) (*entity.DocumentNode, error) {
	content, err := s.readTreeVersions.ReadBlob(ctx, linked.ClonePath(), hash)
	if err != nil {
		return nil, err
	}
	return s.convertSyncDocument.DecodeDocument(ctx, content)
}

func (s syncResolver) store(ctx context.Context, linked entity.LinkedWorkspace, document *entity.DocumentNode) (string, error) {
	content, err := s.convertSyncDocument.EncodeDocument(ctx, document)
	if err != nil {
		return "", err
	}
	return s.storeBlobs.StoreBlob(ctx, linked.ClonePath(), content)
}
