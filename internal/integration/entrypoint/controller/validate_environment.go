package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateListEnvironments(request *dto.ListEnvironments) error {
	return validator.CustomValidator().Struct("ENV-03400", request)
}

func validateGetEnvironment(request *dto.EnvironmentReference) error {
	return validator.CustomValidator().Struct("ENV-13400", request)
}

func validateCreateEnvironment(request *dto.CreateEnvironment) error {
	return validator.CustomValidator().Struct("ENV-11400", request)
}

func validateUpdateEnvironment(request *dto.UpdateEnvironment) error {
	return validator.CustomValidator().Struct("ENV-12400", request)
}

func validateDuplicateEnvironment(request *dto.EnvironmentReference) error {
	return validator.CustomValidator().Struct("ENV-14400", request)
}

func validateSetActiveEnvironment(request *dto.SetActiveEnvironment) error {
	return validator.CustomValidator().Struct("ENV-04400", request)
}

func validateDeleteEnvironment(request *dto.EnvironmentReference) error {
	return validator.CustomValidator().Struct("ENV-15400", request)
}

func validateReorderEnvironment(request *dto.ReorderEnvironment) error {
	return validator.CustomValidator().Struct("ENV-18400", request)
}
