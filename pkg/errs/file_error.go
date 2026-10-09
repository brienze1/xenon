package errs

import (
	"errors"
	"io/fs"
	"net/http"
)

// FileError maps fs.ErrNotExist to 404, fs.ErrPermission to 403 and anything else to 500.
func FileError(err error, code string) BaseError {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return NotFoundError("file not found", code)
	case errors.Is(err, fs.ErrPermission):
		return ClientError("permission denied", code, http.StatusForbidden)
	default:
		return InternalServerError(err, code, "file operation failed")
	}
}
