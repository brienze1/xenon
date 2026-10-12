package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

const repositoryRoot = ""

type lfsDetector struct {
	readBranchHeads  usecase.ReadBranchHeads
	readTreeVersions usecase.ReadTreeVersions
}

func (l lfsDetector) detect(ctx context.Context, linked entity.LinkedWorkspace) (bool, error) {
	heads, err := l.readBranchHeads.ReadHeads(ctx, linked.ClonePath(), linked.Branch())
	if err != nil {
		return false, err
	}
	versions, err := l.readTreeVersions.ReadVersions(ctx, linked.ClonePath(), repositoryRoot, entity.BranchHeads{Local: heads.Local})
	if err != nil {
		return false, err
	}
	attributes := entity.GitAttributes{}
	for path, hash := range versions.Mine {
		if !entity.IsGitAttributesFor(path, linked.Subfolder()) {
			continue
		}
		content, err := l.readTreeVersions.ReadBlob(ctx, linked.ClonePath(), hash)
		if err != nil {
			return false, err
		}
		attributes[path] = string(content)
	}
	return attributes.UsesLfs(), nil
}
