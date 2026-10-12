package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type SyncService struct {
	link     adapter.SyncLinkController
	manage   adapter.SyncManageController
	transfer adapter.SyncTransferController
}

func NewSyncService(link adapter.SyncLinkController, manage adapter.SyncManageController, transfer adapter.SyncTransferController) *SyncService {
	return &SyncService{link: link, manage: manage, transfer: transfer}
}

func (s *SyncService) Status(ctx context.Context, request dto.SyncWorkspaceReference) (*dto.SyncStatus, error) {
	return s.manage.Status(ctx, request)
}

func (s *SyncService) Link(ctx context.Context, request dto.LinkRepository) (*dto.SyncStatus, error) {
	return s.link.Link(ctx, request)
}

func (s *SyncService) Changes(ctx context.Context, request dto.SyncWorkspaceReference) ([]dto.SyncChangeItem, error) {
	return s.manage.Changes(ctx, request)
}

func (s *SyncService) SaveSettings(ctx context.Context, request dto.SaveSyncSettings) error {
	return s.manage.SaveSettings(ctx, request)
}

func (s *SyncService) Commit(ctx context.Context, request dto.CommitSyncChanges) (*dto.SyncStatus, error) {
	return s.transfer.Commit(ctx, request)
}

func (s *SyncService) Pull(ctx context.Context, request dto.SyncWorkspaceReference) (*dto.SyncPullResult, error) {
	return s.transfer.Pull(ctx, request)
}

func (s *SyncService) Push(ctx context.Context, request dto.SyncWorkspaceReference) (*dto.SyncStatus, error) {
	return s.transfer.Push(ctx, request)
}
