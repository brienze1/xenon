package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenWorkspaceSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the workspace "([^"]*)" exists$`, t.theWorkspaceExists)
	ctx.Given(`^the workspace file "([^"]*)" contains$`, t.theWorkspaceFileContains)
	ctx.Given(`^the workspace file "([^"]*)" contains with Windows line endings and a byte order mark$`, t.theWorkspaceFileContainsWithWindowsLineEndingsAndAByteOrderMark)
	ctx.Given(`^the workspace path "([^"]*)" is a symbolic link to "([^"]*)"$`, t.theWorkspacePathIsASymbolicLinkTo)
	ctx.Given(`^the workspace folder "([^"]*)" exists$`, t.theWorkspaceFolderExists)
	ctx.Given(`^the file "([^"]*)" exists with content$`, t.theFileExistsWithContent)
	ctx.Given(`^the workspace path "([^"]*)" is removed$`, t.theWorkspacePathIsRemoved)
	ctx.Given(`^the current time is "([^"]*)"$`, t.theCurrentTimeIs)
	ctx.Given(`^the "([^"]*)" variable is set to the mocked api url$`, t.theVariableIsSetToTheMockedApiUrl)
	ctx.Given(`^the "([^"]*)" variable is set to the mocked tls api url$`, t.theVariableIsSetToTheMockedTlsApiUrl)
	ctx.Given(`^the "([^"]*)" variable is set to a closed port url$`, t.theVariableIsSetToAClosedPortUrl)
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

func (t *testUtils) theWorkspaceFileContainsWithWindowsLineEndingsAndAByteOrderMark(path string, content *godog.DocString) error {
	windows := strings.ReplaceAll(t.substitute(content.Content), "\n", "\r\n") + "\r\n"
	return writeFile(t.workspacePath(path), string(byteOrderMark)+windows)
}

func (t *testUtils) theWorkspacePathIsASymbolicLinkTo(path, target string) error {
	link := t.workspacePath(path)
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return err
	}
	if err := os.Symlink(t.workspacePath(target), link); err != nil {
		if runtime.GOOS == "windows" {
			return godog.ErrSkip
		}
		return err
	}
	return nil
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

func (t *testUtils) theVariableIsSetToTheMockedTlsApiUrl(name string) error {
	t.vars[name] = t.api.TLSURL()
	return nil
}

func (t *testUtils) theVariableIsSetToAClosedPortUrl(name string) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		return err
	}
	t.vars[name] = "http://" + address
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
