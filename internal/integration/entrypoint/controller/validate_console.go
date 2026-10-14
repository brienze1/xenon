package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateListConsole(request *dto.ListConsole) error {
	return validator.CustomValidator().Struct("CON-00400", request)
}

func validateGetConsoleEntry(request *dto.ConsoleEntryReference) error {
	return validator.CustomValidator().Struct("CON-02400", request)
}
