package steps

import (
	"time"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenTimeSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the current time is "([^"]*)"$`, t.theCurrentTimeIs)
	ctx.Step(`^(\d+) seconds? pass(?:es)?$`, t.secondsPass)
}

func (t *testUtils) secondsPass(seconds int) error {
	t.time.Advance(time.Duration(seconds) * time.Second)
	return nil
}

func (t *testUtils) theCurrentTimeIs(value string) error {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return err
	}
	t.time.SetCurrentTime(parsed)
	return nil
}
