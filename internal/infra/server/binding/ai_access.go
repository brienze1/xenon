package binding

import (
	"context"

	wails "github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type AIAccessService struct {
	aiAccess adapter.AIAccessController
}

func NewAIAccessService(aiAccess adapter.AIAccessController) *AIAccessService {
	return &AIAccessService{aiAccess: aiAccess}
}

func (a *AIAccessService) Status(ctx context.Context) (*dto.AIAccessStatus, error) {
	return a.aiAccess.Status(ctx)
}

func (a *AIAccessService) PendingApprovals(ctx context.Context) ([]dto.AIApproval, error) {
	return a.aiAccess.PendingApprovals(ctx)
}

func (a *AIAccessService) AnswerApproval(ctx context.Context, request dto.AIApprovalAnswer) error {
	return a.aiAccess.AnswerApproval(ctx, request)
}

func (a *AIAccessService) ServiceStartup(ctx context.Context, _ wails.ServiceOptions) error {
	_, err := a.aiAccess.Start(ctx)
	return err
}

func (a *AIAccessService) ServiceShutdown() error {
	ctx := context.Background()
	if err := a.aiAccess.DenyApprovals(ctx, dto.AIApprovalDenial{}); err != nil {
		return err
	}
	_, err := a.aiAccess.Stop(ctx)
	return err
}
