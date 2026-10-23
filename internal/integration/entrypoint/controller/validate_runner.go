package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateRunTargetReference(request *dto.RunTargetReference) error {
	return validator.CustomValidator().Struct("RUN-08400", request)
}

func validateSaveRunnerPreferences(request *dto.SaveRunnerPreferences) error {
	return validator.CustomValidator().Struct("RUN-22400", request)
}

func validateRunRowReference(request *dto.RunRowReference) error {
	return validator.CustomValidator().Struct("RUN-10400", request)
}

func validateExportRun(request *dto.ExportRun) error {
	return validator.CustomValidator().Struct("RUN-11400", request)
}

func validateStartRun(request *dto.StartRun) error {
	return validator.CustomValidator().Struct("RUN-09400", request)
}

func validateMCPRun(request *dto.MCPRun) error {
	return validator.CustomValidator().Struct("RUN-24400", request)
}

func validateRunReference(request *dto.RunReference) error {
	return validator.CustomValidator().Struct("RUN-23400", request)
}

func validateWorkspaceReference(request *dto.WorkspaceReference) error {
	return validator.CustomValidator().Struct("RUN-12400", request)
}
