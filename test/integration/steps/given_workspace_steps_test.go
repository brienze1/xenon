package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenWorkspaceSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the workspace "([^"]*)" exists$`, t.theWorkspaceExists)
	ctx.Given(`^the workspace file "([^"]*)" contains$`, t.theWorkspaceFileContains)
	ctx.Given(`^the workspace folder "([^"]*)" exists$`, t.theWorkspaceFolderExists)
	ctx.Given(`^the file "([^"]*)" exists with content$`, t.theFileExistsWithContent)
	ctx.Given(`^the workspace path "([^"]*)" is removed$`, t.theWorkspacePathIsRemoved)
	ctx.Given(`^the current time is "([^"]*)"$`, t.theCurrentTimeIs)
	ctx.Given(`^the "([^"]*)" variable is set to the mocked api url$`, t.theVariableIsSetToTheMockedApiUrl)
	ctx.Given(`^the "([^"]*)" variable is set to "([^"]*)"$`, t.theVariableIsSetTo)
}

func (t *testUtils) theWorkspaceExists(name string) error {
	folder := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	payload, _ := json.Marshal(map[string]string{"name": name, "path": filepath.Join(t.workspaceRoot, folder)})
	if err := t.call(context.Background(), "WorkspaceService", "Create", string(payload)); err != nil {
		return err
	}
	if t.callError != nil {
		return fmt.Errorf("creating workspace %q failed: %s", name, t.callError)
	}
	localId, err := fieldOf(t.response, "local_id")
	if err != nil {
		return err
	}
	t.vars[name+".local_id"] = stringify(localId)
	t.vars[name+".path"] = filepath.Join(t.workspaceRoot, folder)
	return nil
}

func (t *testUtils) theWorkspaceFileContains(path string, content *godog.DocString) error {
	return writeFile(t.workspacePath(path), t.substitute(content.Content))
}

func (t *testUtils) theWorkspaceFolderExists(path string) error {
	return os.MkdirAll(t.workspacePath(path), 0o755)
}

func (t *testUtils) theFileExistsWithContent(path string, content *godog.DocString) error {
	return writeFile(filepath.Join(t.tempRoot, filepath.FromSlash(t.substitute(path))), t.substitute(content.Content))
}

func (t *testUtils) theWorkspacePathIsRemoved(path string) error {
	return os.RemoveAll(t.workspacePath(path))
}

func (t *testUtils) theCurrentTimeIs(value string) error {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return err
	}
	t.time.SetCurrentTime(parsed)
	return nil
}

func (t *testUtils) theVariableIsSetToTheMockedApiUrl(name string) error {
	t.vars[name] = t.api.URL()
	return nil
}

func (t *testUtils) theVariableIsSetTo(name, value string) error {
	t.vars[name] = t.substitute(value)
	return nil
}

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
