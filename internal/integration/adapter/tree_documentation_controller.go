package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type TreeDocumentationController interface {
	GetDocumentation(ctx context.Context, request dto.TreeItemReference) (*dto.Documentation, error)
	PreviewDocumentation(ctx context.Context, request dto.DocumentationInput) (*dto.Documentation, error)
	PlanDocumentation(ctx context.Context, request dto.DocumentationInput) (*dto.DocumentationPlan, error)
	ApplyDocumentation(ctx context.Context, request dto.DocumentationInput) (*dto.DocumentationApplied, error)
}
