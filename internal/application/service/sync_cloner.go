package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

const phaseClone = "clone"

type syncCloner struct {
	findWorkspaceRegistry  usecase.FindWorkspaceRegistry
	inspectWorkspaceFolder usecase.InspectWorkspaceFolder
	cloneRepository        usecase.CloneRepository
	removeWorkspaceFolder  usecase.RemoveWorkspaceFolder
	publishGitEvent        usecase.PublishGitEvent
	lfs                    lfsDetector
}

func (s syncCloner) checkClonePath(ctx context.Context, clonePath string) error {
	registry, err := s.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return err
	}
	if registry.OverlapsAny(clonePath) {
		return usecase.RejectClonePathOverlap(clonePath)
	}
	folder, err := s.inspectWorkspaceFolder.Inspect(ctx, clonePath)
	if err != nil {
		return err
	}
	if !folder.IsEmptyOrAbsent() {
		return usecase.RejectClonePathInUse(clonePath)
	}
	return nil
}

func (s syncCloner) clone(ctx context.Context, connected entity.LinkedWorkspace) error {
	if err := s.publishGitEvent.Progress(ctx, connected.LocalID, phaseClone, 0, 1); err != nil {
		return err
	}
	if err := s.cloneRepository.Clone(ctx, connected.Access, connected.ClonePath(), connected.Settings.Shared()); err != nil {
		return s.discard(ctx, connected.ClonePath(), err)
	}
	usesLfs, err := s.lfs.detect(ctx, connected)
	if err != nil {
		return s.discard(ctx, connected.ClonePath(), err)
	}
	if usesLfs {
		return s.discard(ctx, connected.ClonePath(), usecase.RejectLargeFileStorage(connected.Subfolder()))
	}
	return nil
}

func (s syncCloner) discard(ctx context.Context, clonePath string, cause error) error {
	if err := s.removeWorkspaceFolder.RemoveFolder(ctx, clonePath); err != nil {
		return err
	}
	return cause
}
