package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectMissingImportPassphrase() error {
	return errs.UnauthorizedError("the export's secrets are encrypted; enter its passphrase or import without secrets", "IMP-50401")
}
