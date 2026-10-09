package controller

import (
	"context"
	"errors"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type errorHandler struct{}

func NewErrorHandler() adapter.ErrorHandler {
	return &errorHandler{}
}

func (e *errorHandler) Handle(ctx context.Context, err error, event any) error {
	logger.Error(ctx, err, "Started")

	var customErr errs.BaseError
	if !errors.As(err, &customErr) {
		customErr = errs.InternalServerError(err, "UNK-00500", "Unknown error")
	}

	logger.Error(ctx, customErr, "Finished", event)
	return customErr
}
