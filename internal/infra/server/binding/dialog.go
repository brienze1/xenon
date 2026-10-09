package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type DialogService struct {
	dialog adapter.DialogController
}

func NewDialogService(dialog adapter.DialogController) *DialogService {
	return &DialogService{dialog: dialog}
}

func (d *DialogService) PickFolder(ctx context.Context, request dto.PickFolder) (*dto.PickedPath, error) {
	return d.dialog.PickFolder(ctx, request)
}

func (d *DialogService) PickFile(ctx context.Context, request dto.PickFile) (*dto.PickedPath, error) {
	return d.dialog.PickFile(ctx, request)
}

func (d *DialogService) PickSaveFile(ctx context.Context, request dto.PickSaveFile) (*dto.PickedPath, error) {
	return d.dialog.PickSaveFile(ctx, request)
}
