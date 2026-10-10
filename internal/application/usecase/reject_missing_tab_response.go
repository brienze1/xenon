package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectMissingTabResponse() error {
	return errs.NotFoundError("the tab has no response to save", "EXM-02404")
}
