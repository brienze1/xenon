package controller

import (
	"context"
	"errors"
	"slices"

	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const mcpEditMaxAttempts = 50

var mcpEditConflictCodes = []string{"REQ-02409", "ENV-19409"}

func retryMCPEdit[T any](ctx context.Context, attempt func() (T, error)) (T, error) {
	var result T
	var err error
	for try := 1; try <= mcpEditMaxAttempts; try++ {
		result, err = attempt()
		if !isMCPEditConflict(err) {
			return result, err
		}
		logger.Debug(ctx, "MCP edit lost a revision conflict, retrying", map[string]any{"attempt": try})
	}
	return result, err
}

func isMCPEditConflict(err error) bool {
	var conflict errs.BaseError
	return errors.As(err, &conflict) && slices.Contains(mcpEditConflictCodes, conflict.Code())
}
