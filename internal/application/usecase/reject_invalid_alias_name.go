package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectInvalidAliasName() error {
	return errs.BadRequestError("the alias name must have 1 to 200 characters", "ALS-03400")
}
