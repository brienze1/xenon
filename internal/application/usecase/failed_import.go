package usecase

import (
	"errors"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func FailedImport(item entity.ImportItem, localId string, err error) entity.ImportOutcome {
	failure := entity.ImportFailure{ItemKey: item.Key, Name: item.Name, Message: err.Error()}
	var custom errs.BaseError
	if errors.As(err, &custom) {
		failure.Code = custom.Code()
		failure.Message = custom.Description()
	}
	return item.Failed(localId, failure)
}
