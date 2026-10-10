package steps

import (
	"bytes"
	"fmt"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenResponseContentSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the response should not contain "([^"]*)"$`, t.theResponseShouldNotContain)
}

func (t *testUtils) theResponseShouldNotContain(text string) error {
	if bytes.Contains(t.response, []byte(t.substitute(text))) {
		return fmt.Errorf("the response contains %q: %s", t.substitute(text), t.response)
	}
	return nil
}
