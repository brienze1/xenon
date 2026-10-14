package steps

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenAppSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the app is restarted$`, t.theAppIsRestarted)
}

func (t *testUtils) theAppIsRestarted() error {
	t.restarts++
	restarted := filepath.Join(t.tempRoot, fmt.Sprintf("app-data-restart-%d", t.restarts))
	if err := os.Rename(t.dataDir, restarted); err != nil {
		return err
	}
	t.dataDir = restarted
	t.vars["app_data"] = restarted
	t.closeMcp()
	t.stopMcpServer()
	if err := os.Setenv("XENON_DATA_DIR", restarted); err != nil {
		return err
	}
	t.startMcpServer()
	return nil
}
