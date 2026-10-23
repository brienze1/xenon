package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cucumber/godog"
)

const (
	newmanTimeout = programTimeout
	newmanSwitch  = "XENON_NEWMAN"
	newmanOn      = "on"
)

var (
	newmanFolder  = filepath.Join("..", "..", "..", "tools", "newman")
	newmanRunner  = filepath.Join(newmanFolder, "run.mjs")
	newmanInstall = filepath.Join(newmanFolder, "node_modules", "newman")
)

func (t *testUtils) registerWhenNewmanSteps(ctx *godog.ScenarioContext) {
	ctx.When(`^I send the saved request "([^"]*)" in the workspace "([^"]*)"$`, t.iSendTheSavedRequestInTheWorkspace)
	ctx.When(`^I run Newman on the file "([^"]*)" with the environment file "([^"]*)" in the workspace folder "([^"]*)"$`, t.iRunNewmanOnTheFile)
	ctx.When(`^I run Newman on the file "([^"]*)" with the environment file "([^"]*)" and the data file "([^"]*)" in the workspace folder "([^"]*)"$`, t.iRunNewmanOnTheFileWithTheDataFile)
}

func (t *testUtils) iSendTheSavedRequestInTheWorkspace(id, workspace string) error {
	localId := t.vars[workspace+".local_id"]
	reference, err := json.Marshal(map[string]string{"local_id": localId, "id": t.substitute(id)})
	if err != nil {
		return err
	}
	if err := t.call(context.Background(), "RequestService", "Get", string(reference)); err != nil {
		return err
	}
	if t.callError != nil {
		return fmt.Errorf("opening the request %s failed: %s", id, t.callError)
	}
	payload, err := json.Marshal(map[string]any{"local_id": localId, "tab_id": "saved-" + t.substitute(id), "request": json.RawMessage(t.response)})
	if err != nil {
		return err
	}
	if err := t.call(context.Background(), "RequestService", "Send", string(payload)); err != nil {
		return err
	}
	if t.callError != nil {
		return fmt.Errorf("sending the request %s failed: %s", id, t.callError)
	}
	return nil
}

func (t *testUtils) iRunNewmanOnTheFile(collection, environment, folder string) error {
	return t.runNewman(t.tempPath(collection), t.tempPath(environment), t.workspacePath(folder))
}

func (t *testUtils) iRunNewmanOnTheFileWithTheDataFile(collection, environment, data, folder string) error {
	return t.runNewman(t.tempPath(collection), t.tempPath(environment), t.workspacePath(folder), t.tempPath(data))
}

func (t *testUtils) runNewman(arguments ...string) error {
	_, missing := os.Stat(newmanInstall)
	if missing != nil && os.Getenv(newmanSwitch) != newmanOn {
		return godog.ErrSkip
	}
	if missing != nil {
		return fmt.Errorf("newman is not installed - run make newman: %w", missing)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		return fmt.Errorf("node is required to run Newman: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), newmanTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, node, append([]string{newmanRunner}, arguments...)...)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("newman failed: %w\n%s", err, output)
	}
	return nil
}
