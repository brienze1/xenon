package steps

import (
	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenLogSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the log output is captured$`, t.theLogOutputIsCaptured)
}

func (t *testUtils) theLogOutputIsCaptured() error {
	return t.setEnvironment("LOG_LEVEL", "trace")
}
