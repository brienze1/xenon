package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type UpdateWebservice interface {
	usecase.FindLatestRelease
	usecase.FindReleaseChecksums
	usecase.DownloadUpdatePackage
}
