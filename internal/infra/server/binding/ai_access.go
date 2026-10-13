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

func (a *AIAccessService) ServiceStartup(ctx context.Context, _ wails.ServiceOptions) error {
	_, err := a.aiAccess.Start(ctx)
	return err
}

func (a *AIAccessService) ServiceShutdown() error {
	_, err := a.aiAccess.Stop(context.Background())
	return err
}
