package errors

import "github.com/brienze1/xenon/pkg/errs"

func FailUnknownAICall(err error) error {
	return errs.InternalServerError(err, "UNK-02500", "Unknown error")
}
