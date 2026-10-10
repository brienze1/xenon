package steps

import (
	"fmt"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenLogSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the log output should contain "([^"]*)"$`, t.theLogOutputShouldContain)
	ctx.Then(`^the log output should not contain "([^"]*)"$`, t.theLogOutputShouldNotContain)
}

func (t *testUtils) theLogOutputShouldContain(text string) error {
	if !t.logs.Contains(t.substitute(text)) {
		return fmt.Errorf("the log output does not contain %q", t.substitute(text))
	}
	return nil
}

func (t *testUtils) theLogOutputShouldNotContain(text string) error {
	if t.logs.Contains(t.substitute(text)) {
		return fmt.Errorf("the log output contains %q", t.substitute(text))
	}
	return nil
}
