package binding

import (
	"context"

	wails "github.com/wailsapp/wails/v3/pkg/application"

	serverAdapter "github.com/brienze1/xenon/internal/infra/server/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type AIAccessService struct {
	aiAccess  adapter.AIAccessController
	mcpServer serverAdapter.MCPServer
}

func NewAIAccessService(aiAccess adapter.AIAccessController, mcpServer serverAdapter.MCPServer) *AIAccessService {
	return &AIAccessService{aiAccess: aiAccess, mcpServer: mcpServer}
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
	return a.mcpServer.Start(ctx)
}

func (a *AIAccessService) ServiceShutdown() error {
	return a.mcpServer.Stop(context.Background())
}
