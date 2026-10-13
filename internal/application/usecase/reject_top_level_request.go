package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectTopLevelRequest() error {
	return errs.BadRequestError("A request must be imported into a collection or folder", "IMP-35400")
}
