package steps

import "github.com/cucumber/godog"

func (t *testUtils) registerGivenEventSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the emitted events are forgotten$`, t.theEmittedEventsAreForgotten)
}

func (t *testUtils) theEmittedEventsAreForgotten() error {
	t.events.Reset()
	return nil
}
