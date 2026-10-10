package steps

import (
	"github.com/cucumber/godog"

	"github.com/brienze1/xenon/pkg/logger"
)

func (t *testUtils) registerGivenLogSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the log output is captured$`, t.theLogOutputIsCaptured)
}

func (t *testUtils) theLogOutputIsCaptured() error {
	t.logs.Reset()
	logger.SetOutput(t.logs)
	return t.setEnvironment("LOG_LEVEL", "trace")
}
