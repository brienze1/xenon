package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectStaleEnvironmentWrite() error {
	return errs.ConflictError("the environment changed on disk since it was opened", "ENV-19409")
}
