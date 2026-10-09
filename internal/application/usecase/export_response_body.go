package usecase

import "context"

type ExportResponseBody interface {
	Export(ctx context.Context, tabId, path string) error
}
