package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectUnknownEnvironment() error {
	return errs.BadRequestError("the environment does not exist in this workspace", "ALS-10400")
}
