package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateListHistory(request *dto.ListHistory) error {
	return validator.CustomValidator().Struct("HIS-00400", request)
}

func validateGetHistory(request *dto.HistoryReference) error {
	return validator.CustomValidator().Struct("HIS-01400", request)
}

func validateClearHistory(request *dto.ClearHistory) error {
	return validator.CustomValidator().Struct("HIS-02400", request)
}
