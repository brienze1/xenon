package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectReservedAliasName() error {
	return errs.BadRequestError(`"Default" is reserved; choose another alias name`, "ALS-04400")
}
