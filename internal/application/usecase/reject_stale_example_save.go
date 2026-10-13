package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectStaleExampleSave() error {
	return errs.ConflictError("the request changed on disk since it was opened", "REQ-02409")
}
