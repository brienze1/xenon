package service

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

func localVariablesOrDefaults(ctx context.Context, localId string, local *entity.LocalVariables, err error) (*entity.LocalVariables, error) {
	switch {
	case err == nil:
		return local, nil
	case errs.IsNotFoundError(err):
		logger.Info(ctx, "No local variables saved yet", map[string]any{"localId": localId})
	case errs.IsUnprocessableEntityError(err):
		logger.Warn(ctx, err, "variables.json is not valid JSON, starting from no local state", map[string]any{"localId": localId})
	default:
		return nil, err
	}
	defaults := entity.NewLocalVariables()
	return &defaults, nil
}
