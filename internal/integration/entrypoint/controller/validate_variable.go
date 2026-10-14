package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateGetVariableTable(request *dto.VariableTableReference) error {
	return validator.CustomValidator().Struct("VAR-05400", request)
}

func validateSaveVariableTable(request *dto.SaveVariableTable) error {
	return validator.CustomValidator().Struct("VAR-08400", request)
}

func validateSetCurrentValue(request *dto.SetCurrentValue) error {
	return validator.CustomValidator().Struct("VAR-09400", request)
}

func validateResetVariables(request *dto.VariableKeys) error {
	return validator.CustomValidator().Struct("VAR-10400", request)
}

func validatePersistVariables(request *dto.VariableKeys) error {
	return validator.CustomValidator().Struct("VAR-11400", request)
}

func validateAddVariable(request *dto.AddVariable) error {
	return validator.CustomValidator().Struct("VAR-12400", request)
}

func validateInspectVariables(request *dto.InspectVariables) error {
	return validator.CustomValidator().Struct("VAR-15400", request)
}
