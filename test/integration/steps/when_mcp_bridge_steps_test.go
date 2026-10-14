package steps

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cucumber/godog"
)

const (
	bridgeTimeout = 20 * time.Second
	bridgePackage = "internal/infra/mcpbridge"
)

var (
	bridgeBinaryOnce sync.Once
	bridgeBinary     string
	bridgeBuildError error
)

type bridgeRun struct {
	stdout   string
	stderr   string
	exitCode int
}

func (t *testUtils) registerWhenMcpBridgeSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I run the stdio bridge with the input$`, t.iRunTheStdioBridgeWithTheInput)
}

func (t *testUtils) iRunTheStdioBridgeWithTheInput(input *godog.DocString) error {
	binary, err := buildBridge()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), bridgeTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, binary, "mcp")
	command.Env = append(os.Environ(), "XENON_DATA_DIR="+t.dataDir)
	command.Stdin = strings.NewReader(t.substitute(input.Content) + "\n")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	runErr := command.Run()
	if ctx.Err() != nil {
		return fmt.Errorf("the stdio bridge did not exit within %s; stderr: %s", bridgeTimeout, stderr.String())
	}
	run := &bridgeRun{stdout: stdout.String(), stderr: stderr.String()}
	var exitErr *exec.ExitError
	switch {
	case errors.As(runErr, &exitErr):
		run.exitCode = exitErr.ExitCode()
	case runErr != nil:
		return runErr
	}
	t.bridge = run
	return nil
}

func buildBridge() (string, error) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(bridgePackage))); err != nil {
		return "", fmt.Errorf("the stdio bridge is not implemented yet (%s is missing)", bridgePackage)
	}
	bridgeBinaryOnce.Do(func() {
		folder, err := os.MkdirTemp("", "xenon-bridge-")
		if err != nil {
			bridgeBuildError = err
			return
		}
		bridgeBinary = filepath.Join(folder, "xenon")
		if runtime.GOOS == "windows" {
			bridgeBinary += ".exe"
		}
		build := exec.Command("go", "build", "-o", bridgeBinary, "./cmd/xenon")
		build.Dir = root
		if output, err := build.CombinedOutput(); err != nil {
			bridgeBuildError = fmt.Errorf("building ./cmd/xenon: %w: %s", err, output)
		}
	})
	return bridgeBinary, bridgeBuildError
}
