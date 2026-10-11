package steps

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/cucumber/godog"
)

const (
	programTimeout = 3 * time.Minute
	programsSwitch = "XENON_SNIPPET_PROGRAMS"
	programsOff    = "off"
)

var errRuntimeMissing = errors.New("runtime not installed")

type program struct {
	file string
	tool string
	args []string
}

var programs = map[string]program{
	"curl":   {file: "snippet.sh", tool: "bash"},
	"node":   {file: "snippet.js", tool: "node"},
	"go":     {file: "snippet.go", tool: "go", args: []string{"run"}},
	"python": {file: "snippet.py"},
}

func (t *testUtils) registerWhenProgramSteps(ctx *godog.ScenarioContext) {
	ctx.When(`^I run the response field "([^"]*)" with "(curl|node|go|python)" in the workspace folder "([^"]*)"$`, t.iRunTheResponseFieldWithInTheWorkspaceFolder)
}

func (t *testUtils) iRunTheResponseFieldWithInTheWorkspaceFolder(field, name, folder string) error {
	value, err := t.responseField(field)
	if err != nil {
		return err
	}
	if os.Getenv(programsSwitch) == programsOff {
		return godog.ErrSkip
	}
	tool, err := programTool(name)
	if errors.Is(err, errRuntimeMissing) && os.Getenv("CI") == "" {
		return godog.ErrSkip
	}
	if err != nil {
		return err
	}
	directory, err := os.MkdirTemp(t.tempRoot, "program-")
	if err != nil {
		return err
	}
	source := stringify(value)
	file := filepath.Join(directory, programs[name].file)
	if err := os.WriteFile(file, []byte(source), 0o600); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), programTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, tool, append(append([]string{}, programs[name].args...), filepath.ToSlash(file))...)
	command.Dir = t.workspacePath(folder)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("running the %s program failed: %w\n%s\n--- program ---\n%s", name, err, output, source)
	}
	return nil
}

func programTool(name string) (string, error) {
	if name == "python" {
		return pythonWithRequests()
	}
	if name == "curl" {
		if _, err := exec.LookPath("curl"); err != nil {
			return "", fmt.Errorf("curl: %w", errRuntimeMissing)
		}
	}
	if name == "go" {
		return goTool(), nil
	}
	tool, err := exec.LookPath(programs[name].tool)
	if err != nil {
		return "", fmt.Errorf("%s: %w", programs[name].tool, errRuntimeMissing)
	}
	return tool, nil
}

func goTool() string {
	if tool, err := exec.LookPath("go"); err == nil {
		return tool
	}
	return filepath.Join(runtime.GOROOT(), "bin", "go")
}

func pythonWithRequests() (string, error) {
	candidates := []string{os.Getenv("XENON_SNIPPET_PYTHON"), "python3", "python"}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		tool, err := exec.LookPath(candidate)
		if err != nil {
			continue
		}
		if exec.Command(tool, "-c", "import requests").Run() == nil {
			return tool, nil
		}
	}
	return "", fmt.Errorf("python with the requests package: %w", errRuntimeMissing)
}
