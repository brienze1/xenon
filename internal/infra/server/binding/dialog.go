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
