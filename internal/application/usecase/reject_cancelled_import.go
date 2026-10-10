package usecase

import "github.com/brienze1/xenon/pkg/errs"

const cancelledStatus = 499

func RejectCancelledImport() error {
	return errs.ClientError("the import was cancelled before anything was written", "IMP-10499", cancelledStatus)
}
