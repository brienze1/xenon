package usecase

import "context"

type SaveRunExport interface {
	SaveRunExport(ctx context.Context, path string, data []byte) error
}
