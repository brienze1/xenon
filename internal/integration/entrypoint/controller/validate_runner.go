package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateStartRun(request *dto.StartRun) error {
	return validator.CustomValidator().Struct("RUN-09400", request)
}

func validateRunReference(request *dto.RunReference) error {
	return validator.CustomValidator().Struct("RUN-10400", request)
}

func validateWorkspaceReference(request *dto.WorkspaceReference) error {
	return validator.CustomValidator().Struct("RUN-12400", request)
}
