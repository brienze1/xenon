package steps

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenMcpBridgeSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the bridge exit code should be (\d+)$`, t.theBridgeExitCodeShouldBe)
	ctx.Then(`^the bridge output should contain "([^"]*)"$`, t.theBridgeOutputShouldContain)
	ctx.Then(`^the bridge output should not contain "([^"]*)"$`, t.theBridgeOutputShouldNotContain)
	ctx.Then(`^the bridge output should match "([^"]*)"$`, t.theBridgeOutputShouldMatch)
	ctx.Then(`^the bridge error output should contain "([^"]*)"$`, t.theBridgeErrorOutputShouldContain)
}

func (t *testUtils) bridgeResult() (*bridgeRun, error) {
	if t.bridge == nil {
		return nil, fmt.Errorf("the stdio bridge has not been run")
	}
	return t.bridge, nil
}

func (t *testUtils) theBridgeExitCodeShouldBe(code int) error {
	run, err := t.bridgeResult()
	if err != nil {
		return err
	}
	if run.exitCode != code {
		return fmt.Errorf("expected exit code %d, got %d; stderr: %s", code, run.exitCode, run.stderr)
	}
	return nil
}

func (t *testUtils) theBridgeOutputShouldContain(text string) error {
	run, err := t.bridgeResult()
	if err != nil {
		return err
	}
	if !strings.Contains(run.stdout, t.substitute(text)) {
		return fmt.Errorf("the bridge output does not contain %q: %s", t.substitute(text), run.stdout)
	}
	return nil
}

func (t *testUtils) theBridgeOutputShouldNotContain(text string) error {
	run, err := t.bridgeResult()
	if err != nil {
		return err
	}
	if strings.Contains(run.stdout, t.substitute(text)) {
		return fmt.Errorf("the bridge output contains %q: %s", t.substitute(text), run.stdout)
	}
	return nil
}

func (t *testUtils) theBridgeErrorOutputShouldContain(text string) error {
	run, err := t.bridgeResult()
	if err != nil {
		return err
	}
	if !strings.Contains(run.stderr, t.substitute(text)) {
		return fmt.Errorf("the bridge error output does not contain %q: %s", t.substitute(text), run.stderr)
	}
	return nil
}

func (t *testUtils) theBridgeOutputShouldMatch(pattern string) error {
	run, err := t.bridgeResult()
	if err != nil {
		return err
	}
	expression, err := regexp.Compile(t.substitute(pattern))
	if err != nil {
		return err
	}
	if !expression.MatchString(run.stdout) {
		return fmt.Errorf("the bridge output does not match %q: %s", pattern, run.stdout)
	}
	return nil
}
