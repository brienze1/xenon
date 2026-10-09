package controller

import (
	"context"
	"fmt"
	"runtime"

	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

func handlePanic(ctx context.Context, event any, recovered any) error {
	var message string
	switch e := recovered.(type) {
	case string:
		message = fmt.Sprint("recovered (string) panic: ", e)
	case runtime.Error:
		message = fmt.Sprint("recovered (runtime.Error) panic: ", e.Error())
	case error:
		message = fmt.Sprint("recovered (error) panic: ", e.Error())
	default:
		message = fmt.Sprint("recovered (default) panic: ", e)
	}

	err := errs.PanicError(message)
	logger.Error(ctx, err, "Panic occurred", event)
	return err
}
