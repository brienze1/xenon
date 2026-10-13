package usecase

import "github.com/brienze1/xenon/pkg/errs"

func RejectInvalidExportCombination() error {
	return errs.BadRequestError("this item cannot be exported in that format and shape (secrets can only be included in Xenon YAML)", "EXP-00400")
}

func RejectInvalidExportPassphrase() error {
	return errs.BadRequestError("the passphrase must have at least 8 characters and both entries must match", "EXP-01400")
}
