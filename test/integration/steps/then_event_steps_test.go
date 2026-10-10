package steps

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenEventSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the event "([^"]*)" should have been emitted (\d+) times$`, t.theEventShouldHaveBeenEmittedTimes)
	ctx.Then(`^the "([^"]*)" events should have been emitted at most (\d+) ms apart$`, t.theEventsShouldHaveBeenEmittedAtMostApart)
	ctx.Then(`^the (\d+) "([^"]*)" event should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theEventShouldContainTheFieldEqualTo)
}

func (t *testUtils) theEventShouldHaveBeenEmittedTimes(name string, times int) error {
	if count := t.events.Count(name); count != times {
		return fmt.Errorf("event %s: expected %d emissions, got %d", name, times, count)
	}
	return nil
}

func (t *testUtils) theEventShouldContainTheFieldEqualTo(index int, name, path, expected string) error {
	data, found := t.events.Event(name, index)
	if !found {
		return fmt.Errorf("event %s #%d was not emitted", name, index)
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}
	value, err := fieldOf(encoded, path)
	if err != nil {
		return err
	}
	return expectEqual(stringify(value), t.substitute(expected), path)
}

func (t *testUtils) theEventsShouldHaveBeenEmittedAtMostApart(name string, milliseconds int) error {
	limit := time.Duration(milliseconds) * time.Millisecond
	for index, gap := range t.events.Gaps(name) {
		if gap > limit {
			return fmt.Errorf("events %s #%d and #%d were %s apart, expected at most %s", name, index+1, index+2, gap, limit)
		}
	}
	return nil
}
