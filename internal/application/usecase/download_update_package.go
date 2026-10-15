package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type DownloadUpdatePackage interface {
	// DownloadUpdatePackage keeps the package only when it matches its checksum and returns where it was kept.
	DownloadUpdatePackage(ctx context.Context, asset entity.ReleaseAsset, checksums entity.ReleaseChecksums, progress func(received, total int64)) (string, error)
}
