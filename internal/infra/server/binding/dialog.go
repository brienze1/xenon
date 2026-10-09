package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/infra/dependency/external"
)

type DialogService struct {
	dialog adapter.Dialog
}

func NewDialogService() *DialogService {
	return &DialogService{dialog: external.GetDialog()}
}

func (d *DialogService) PickFolder(_ context.Context, request PickFolderRequest) (*PickedPath, error) {
	path, err := d.dialog.PickFolder(request.Title)
	if err != nil {
		return nil, err
	}
	return &PickedPath{Path: path}, nil
}

func (d *DialogService) PickFile(_ context.Context, request PickFileRequest) (*PickedPath, error) {
	path, err := d.dialog.PickFile(request.Title)
	if err != nil {
		return nil, err
	}
	return &PickedPath{Path: path}, nil
}

func (d *DialogService) PickSaveFile(_ context.Context, request PickSaveFileRequest) (*PickedPath, error) {
	path, err := d.dialog.PickSaveFile(request.Title, request.SuggestedName)
	if err != nil {
		return nil, err
	}
	return &PickedPath{Path: path}, nil
}
