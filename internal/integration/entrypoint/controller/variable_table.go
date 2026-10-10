package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type variableTable struct {
	errorHandler               adapter.ErrorHandler
	manageVariableTableService appAdapter.ManageVariableTableService
}

func NewVariableTable(errorHandler adapter.ErrorHandler, manageVariableTableService appAdapter.ManageVariableTableService) adapter.VariableTableController {
	return &variableTable{errorHandler: errorHandler, manageVariableTableService: manageVariableTableService}
}

func (v *variableTable) GetTable(bindingCtx context.Context, request dto.VariableTableReference) (response *dto.VariableTable, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, v.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("VAR-05400", &request)
	if err != nil {
		return nil, err
	}

	table, err := v.manageVariableTableService.GetTable(ctx, request.LocalId, request.ScopeRef())
	if err != nil {
		return nil, err
	}
	result := dto.NewVariableTable(*table)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (v *variableTable) SaveTable(bindingCtx context.Context, request dto.SaveVariableTable) (response *dto.VariableTable, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, v.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("VAR-08400", &request)
	if err != nil {
		return nil, err
	}

	table, err := v.manageVariableTableService.SaveTable(ctx, request.LocalId, request.ScopeRef(), request.ToEntity(), request.ConfirmReveal)
	if err != nil {
		return nil, err
	}
	result := dto.NewVariableTable(*table)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (v *variableTable) SetCurrent(bindingCtx context.Context, request dto.SetCurrentValue) (response *dto.VariableRow, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, v.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("VAR-09400", &request)
	if err != nil {
		return nil, err
	}

	entry, err := v.manageVariableTableService.SetCurrent(ctx, request.LocalId, request.ScopeRef(), request.Key, request.Value)
	if err != nil {
		return nil, err
	}
	result := dto.NewVariableRow(*entry)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (v *variableTable) Reset(bindingCtx context.Context, request dto.VariableKeys) (response *dto.VariableTable, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, v.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("VAR-10400", &request)
	if err != nil {
		return nil, err
	}

	table, err := v.manageVariableTableService.Reset(ctx, request.LocalId, request.ScopeRef(), request.Keys)
	if err != nil {
		return nil, err
	}
	result := dto.NewVariableTable(*table)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (v *variableTable) Persist(bindingCtx context.Context, request dto.VariableKeys) (response *dto.VariableTable, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, v.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("VAR-11400", &request)
	if err != nil {
		return nil, err
	}

	table, err := v.manageVariableTableService.Persist(ctx, request.LocalId, request.ScopeRef(), request.Keys)
	if err != nil {
		return nil, err
	}
	result := dto.NewVariableTable(*table)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (v *variableTable) AddVariable(bindingCtx context.Context, request dto.AddVariable) (response *dto.VariableRow, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, v.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("VAR-12400", &request)
	if err != nil {
		return nil, err
	}

	entry, err := v.manageVariableTableService.AddVariable(ctx, request.LocalId, request.ScopeRef(), request.ToEntity(), request.Value)
	if err != nil {
		return nil, err
	}
	result := dto.NewVariableRow(*entry)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
