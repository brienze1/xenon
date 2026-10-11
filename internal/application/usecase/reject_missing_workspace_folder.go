package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectMissingWorkspaceFolder() error {
	return errs.BadRequestError("choose an empty or new folder for the imported workspace", "IMP-43400")
}
