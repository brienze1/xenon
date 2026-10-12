package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

type syncSnapshot struct {
	head     map[string]string
	worktree entity.SyncedFiles
	changes  entity.ChangeSet
	heads    entity.BranchHeads
}

type syncChanges struct {
	readBranchHeads   usecase.ReadBranchHeads
	readTreeVersions  usecase.ReadTreeVersions
	identifySyncItems usecase.IdentifySyncItems
}

func newSyncChanges(readBranchHeads usecase.ReadBranchHeads, readTreeVersions usecase.ReadTreeVersions, identifySyncItems usecase.IdentifySyncItems) syncChanges {
	return syncChanges{readBranchHeads: readBranchHeads, readTreeVersions: readTreeVersions, identifySyncItems: identifySyncItems}
}

func (s syncChanges) snapshot(ctx context.Context, linked entity.LinkedWorkspace) (*syncSnapshot, error) {
	heads, err := s.readBranchHeads.ReadHeads(ctx, linked.ClonePath(), linked.Branch())
	if err != nil {
		return nil, err
	}
	head, err := s.tree(ctx, linked, heads.Local)
	if err != nil {
		return nil, err
	}
	files, err := s.readTreeVersions.ReadWorktree(ctx, linked.ClonePath(), linked.Subfolder())
	if err != nil {
		return nil, err
	}
	worktree := entity.NewSyncedFiles(files)
	pathChanges := worktree.ChangesFrom(head)
	present, deleted := make([]string, 0, len(pathChanges)), map[string]string{}
	for _, change := range pathChanges {
		if change.IsDeletion() {
			deleted[change.Path] = head[change.Path]
			continue
		}
		present = append(present, change.Path)
	}
	refs, err := s.identify(ctx, linked, present, deleted)
	if err != nil {
		return nil, err
	}
	changes := entity.NewChangeSet(usecase.MapPathsToItems(pathChanges, refs))
	return &syncSnapshot{heads: *heads, head: head, worktree: worktree, changes: changes}, nil
}

func (s syncChanges) tree(ctx context.Context, linked entity.LinkedWorkspace, commit string) (map[string]string, error) {
	versions, err := s.readTreeVersions.ReadVersions(ctx, linked.ClonePath(), linked.Subfolder(), entity.BranchHeads{Local: commit})
	if err != nil {
		return nil, err
	}
	return versions.Mine, nil
}

func (s syncChanges) identify(ctx context.Context, linked entity.LinkedWorkspace, present []string, blobs map[string]string) (map[string]entity.ItemRef, error) {
	contents, err := s.readTreeVersions.ReadWorktreeFiles(ctx, linked.ClonePath(), linked.Subfolder(), present)
	if err != nil {
		return nil, err
	}
	for path, hash := range blobs {
		content, err := s.readTreeVersions.ReadBlob(ctx, linked.ClonePath(), hash)
		if err != nil {
			return nil, err
		}
		contents[path] = content
	}
	return s.identifySyncItems.IdentifyItems(ctx, contents), nil
}
