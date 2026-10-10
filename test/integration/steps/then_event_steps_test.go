package steps

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenEventSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the event "([^"]*)" should have been emitted (\d+) times$`, t.theEventShouldHaveBeenEmittedTimes)
	ctx.Step(`^the "([^"]*)" events should have been emitted at most (\d+) ms apart$`, t.theEventsShouldHaveBeenEmittedAtMostApart)
	ctx.Then(`^the (\d+) "([^"]*)" event should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theEventShouldContainTheFieldEqualTo)
	ctx.Then(`^the event "([^"]*)" should have been emitted at most (\d+) times$`, t.theEventShouldHaveBeenEmittedAtMostTimes)
	ctx.Step(`^the "([^"]*)" events should have (\d+) items in the field "([^"]*)" within (\d+) ms$`, t.theEventsShouldHaveItemsInTheFieldWithin)
	ctx.Then(`^an? "([^"]*)" event with the field "([^"]*)" equal to "([^"]*)" should have been emitted within (\d+) ms$`, t.anEventWithTheFieldEqualToShouldHaveBeenEmittedWithin)
	ctx.Then(`^the last "([^"]*)" event should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theLastEventShouldContainTheFieldEqualTo)
}

const eventPollInterval = 5 * time.Millisecond

func (t *testUtils) theEventShouldHaveBeenEmittedTimes(name string, times int) error {
	if count := t.events.Count(name); count != times {
		return fmt.Errorf("event %s: expected %d emissions, got %d", name, times, count)
	}
	return nil
}

func (t *testUtils) theEventShouldContainTheFieldEqualTo(index int, name, path, expected string) error {
	value, err := t.eventField(name, index, path)
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

func (t *testUtils) theEventShouldHaveBeenEmittedAtMostTimes(name string, times int) error {
	if count := t.events.Count(name); count > times {
		return fmt.Errorf("event %s: expected at most %d emissions, got %d", name, times, count)
	}
	return nil
}

func (t *testUtils) theEventsShouldHaveItemsInTheFieldWithin(name string, items int, path string, milliseconds int) error {
	total := 0
	err := waitFor(milliseconds, func() (bool, error) {
		total = 0
		for index := 1; index <= t.events.Count(name); index++ {
			value, err := t.eventField(name, index, path)
			if err != nil {
				return false, err
			}
			list, _ := value.([]any)
			total += len(list)
		}
		return total == items, nil
	})
	if err != nil {
		return err
	}
	if total != items {
		return fmt.Errorf("events %s: expected %d items in %s within %d ms, got %d", name, items, path, milliseconds, total)
	}
	return nil
}

func (t *testUtils) anEventWithTheFieldEqualToShouldHaveBeenEmittedWithin(name, path, expected string, milliseconds int) error {
	expected = t.substitute(expected)
	found := false
	err := waitFor(milliseconds, func() (bool, error) {
		for index := 1; index <= t.events.Count(name); index++ {
			if value, err := t.eventField(name, index, path); err == nil && stringify(value) == expected {
				found = true
			}
		}
		return found, nil
	})
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("no %s event with %s equal to %q within %d ms", name, path, expected, milliseconds)
	}
	return nil
}

func (t *testUtils) theLastEventShouldContainTheFieldEqualTo(name, path, expected string) error {
	return t.theEventShouldContainTheFieldEqualTo(t.events.Count(name), name, path, expected)
}

func (t *testUtils) eventField(name string, index int, path string) (any, error) {
	data, found := t.events.Event(name, index)
	if !found {
		return nil, fmt.Errorf("event %s #%d was not emitted", name, index)
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return fieldOf(encoded, path)
}

func waitFor(milliseconds int, done func() (bool, error)) error {
	deadline := time.Now().Add(time.Duration(milliseconds) * time.Millisecond)
	for {
		finished, err := done()
		if err != nil || finished || time.Now().After(deadline) {
			return err
		}
		time.Sleep(eventPollInterval)
	}
}
