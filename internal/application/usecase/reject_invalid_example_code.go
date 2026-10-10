package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectInvalidExampleCode() error {
	return errs.BadRequestError("the response status code must be between 100 and 599", "EXM-04400")
}
