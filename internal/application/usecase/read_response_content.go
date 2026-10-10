package usecase

import (
	"io"

	"github.com/brienze1/xenon/pkg/errs"
)

func ReadResponseContent(content io.ReadCloser) ([]byte, error) {
	body, err := io.ReadAll(content)
	closeErr := content.Close()
	if err != nil {
		return nil, errs.FileError(err, "EXM-12500")
	}
	if closeErr != nil {
		return nil, errs.FileError(closeErr, "EXM-13500")
	}
	return body, nil
}
