package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

type RenderSyncItem interface {
	RenderItem(ctx context.Context, kind enums.ItemKind, path string, content []byte) entity.ConflictVersion
	// RenderNode returns nil for an absent node.
	RenderNode(ctx context.Context, node *entity.DocumentNode) (*string, error)
}
