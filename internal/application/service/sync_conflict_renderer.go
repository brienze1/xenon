package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

type syncConflictRenderer struct {
	readTreeVersions usecase.ReadTreeVersions
	renderSyncItem   usecase.RenderSyncItem
}

func (s syncConflictRenderer) render(ctx context.Context, linked entity.LinkedWorkspace, conflict entity.ConflictItem, worktree entity.SyncedFiles) (*entity.ConflictItem, error) {
	versions := make([]*entity.ConflictVersion, 0, 3)
	for _, hash := range []string{conflict.Base, conflict.Mine, conflict.Theirs} {
		version, err := s.version(ctx, linked, conflict, hash, worktree)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	fields := make([]entity.FieldConflict, 0, len(conflict.Fields))
	for _, field := range conflict.Fields {
		mine, err := s.renderSyncItem.RenderNode(ctx, field.Mine)
		if err != nil {
			return nil, err
		}
		theirs, err := s.renderSyncItem.RenderNode(ctx, field.Theirs)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field.WithTexts(mine, theirs))
	}
	rendered := conflict.WithVersions(entity.ConflictVersions{Base: versions[0], Mine: versions[1], Theirs: versions[2]}).WithFields(fields)
	return &rendered, nil
}

func (s syncConflictRenderer) version(ctx context.Context, linked entity.LinkedWorkspace, conflict entity.ConflictItem, hash string, worktree entity.SyncedFiles) (*entity.ConflictVersion, error) {
	if hash == "" {
		return nil, nil
	}
	content, err := s.content(ctx, linked, conflict.Path, hash, worktree)
	if err != nil {
		return nil, err
	}
	version := s.renderSyncItem.RenderItem(ctx, conflict.ItemKind, conflict.Path, content)
	return &version, nil
}

func (s syncConflictRenderer) content(ctx context.Context, linked entity.LinkedWorkspace, path, hash string, worktree entity.SyncedFiles) ([]byte, error) {
	if !worktree.Holds(path, hash) {
		return s.readTreeVersions.ReadBlob(ctx, linked.ClonePath(), hash)
	}
	files, err := s.readTreeVersions.ReadWorktreeFiles(ctx, linked.ClonePath(), linked.Subfolder(), []string{path})
	if err != nil {
		return nil, err
	}
	return files[path], nil
}
