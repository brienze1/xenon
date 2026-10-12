package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageDocumentationService interface {
	Get(ctx context.Context, localId, id string) (*entity.ApiDocumentation, error)
	Preview(ctx context.Context, text string) (*entity.ApiDocumentation, error)
	Plan(ctx context.Context, localId, id, text string) (*entity.DocumentationPlan, error)
	Apply(ctx context.Context, localId, id, text string, confirmDeletions bool) (*entity.DocumentationResult, error)
}
