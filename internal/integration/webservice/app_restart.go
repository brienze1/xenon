package webservice

import (
	"context"
	"os"
	"strconv"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const restartAfterPidVariable = "XENON_RESTART_AFTER_PID"

type appRestartWebservice struct {
	appProcess adapter.AppProcess
}

func NewAppRestartWebservice(appProcess adapter.AppProcess) adapter.AppRestartWebservice {
	return &appRestartWebservice{appProcess: appProcess}
}

func (a *appRestartWebservice) RestartApp(ctx context.Context, executable string) error {
	logger.Debug(ctx, "Started", map[string]any{"executable": executable})

	if err := a.appProcess.Spawn(executable, map[string]string{restartAfterPidVariable: strconv.Itoa(os.Getpid())}); err != nil {
		return errs.ServiceUnavailableError(err, "UPD-07503", "The updated Xenon could not be started")
	}
	a.appProcess.Quit()

	logger.Debug(ctx, "Finished", map[string]any{"executable": executable})
	return nil
}
