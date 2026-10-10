package steps

import (
	"time"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenTimeSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the current time is "([^"]*)"$`, t.theCurrentTimeIs)
}

func (t *testUtils) theCurrentTimeIs(value string) error {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return err
	}
	t.time.SetCurrentTime(parsed)
	return nil
}
