package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectInvalidExampleName() error {
	return errs.BadRequestError("the example name must have 1 to 200 characters", "EXM-03400")
}
