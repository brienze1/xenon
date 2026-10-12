package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

type syncMerger struct {
	readTreeVersions    usecase.ReadTreeVersions
	storeBlobs          usecase.StoreBlobs
	convertSyncDocument usecase.ConvertSyncDocument
}

func (s syncMerger) merge(ctx context.Context, linked entity.LinkedWorkspace, heads entity.BranchHeads, resolutions map[string]map[string]enums.ConflictResolution) (*pullTarget, error) {
	versions, err := s.readTreeVersions.ReadVersions(ctx, linked.ClonePath(), linked.Subfolder(), heads)
	if err != nil {
		return nil, err
	}
	plan := usecase.PlanThreeWay(versions.Base, versions.Mine, versions.Theirs)
	target := &pullTarget{conflicts: plan.Conflicts(), fields: map[string][]entity.FieldConflict{}}
	merged := map[string]string{}
	for _, structural := range plan.StructuralMerges() {
		result, err := s.mergeFile(ctx, linked, structural, resolutions[structural.Path])
		if err != nil {
			return nil, err
		}
		if result.HasConflicts() {
			target.conflicts = append(target.conflicts, structural)
			target.fields[structural.Path] = result.FieldConflicts
			continue
		}
		content, err := s.convertSyncDocument.EncodeDocument(ctx, result.Result)
		if err != nil {
			return nil, err
		}
		if merged[structural.Path], err = s.storeBlobs.StoreBlob(ctx, linked.ClonePath(), content); err != nil {
			return nil, err
		}
	}
	target.tree = plan.Merged(merged)
	return target, nil
}

func (s syncMerger) mergeFile(ctx context.Context, linked entity.LinkedWorkspace, plan entity.PathPlan, resolutions map[string]enums.ConflictResolution) (*entity.StructuralMerge, error) {
	documents := make([]*entity.DocumentNode, 0, 3)
	for _, hash := range []string{plan.Base, plan.Mine, plan.Theirs} {
		document, err := s.document(ctx, linked, hash)
		if err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}
	result := usecase.MergeStructuralFile(plan.StructuralKind(), documents[0], documents[1], documents[2], resolutions)
	return &result, nil
}

func (s syncMerger) document(ctx context.Context, linked entity.LinkedWorkspace, hash string) (*entity.DocumentNode, error) {
	if hash == "" {
		return nil, nil
	}
	content, err := s.readTreeVersions.ReadBlob(ctx, linked.ClonePath(), hash)
	if err != nil {
		return nil, err
	}
	return s.convertSyncDocument.DecodeDocument(ctx, content)
}
