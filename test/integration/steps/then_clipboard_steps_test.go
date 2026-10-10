package steps

import (
	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenClipboardSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the clipboard should contain "([^"]*)"$`, t.theClipboardShouldContain)
}

func (t *testUtils) theClipboardShouldContain(text string) error {
	return expectEqual(t.clipboard.Text(), t.substitute(text), "clipboard")
}
