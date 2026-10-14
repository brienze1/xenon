package controller

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateAIArguments(request any) error {
	return validator.CustomValidator().Struct(entity.AIInvalidArgumentsCode, request)
}

func validateAIGuard(request *dto.AIGuard) error {
	return validator.CustomValidator().Struct("MCP-09400", request)
}

func validateAIChanges(request *dto.AIChanges) error {
	return validator.CustomValidator().Struct("MCP-10400", request)
}

func validateAIActivityRecord(request *dto.AIActivityRecord) error {
	return validator.CustomValidator().Struct("MCP-11400", request)
}
