package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectTakenAliasName() error {
	return errs.ConflictError("this request already has an alias with that name", "ALS-05409")
}
