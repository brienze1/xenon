package steps

import (
	"fmt"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenApiSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the header "([^"]*)" equal to "([^"]*)"$`, t.theRequestShouldHaveTheHeaderEqualTo)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the query "([^"]*)" equal to "([^"]*)"$`, t.theRequestShouldHaveTheQueryEqualTo)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the body equal to$`, t.theRequestShouldHaveTheBodyEqualTo)
	ctx.Then(`^the "([^"]*)" request to "([^"]*)" should have been received (\d+) times$`, t.theRequestShouldHaveBeenReceivedTimes)
}

func (t *testUtils) theRequestShouldHaveTheHeaderEqualTo(index int, method, path, key, expected string) error {
	value, found := t.api.GetRequestHeader(index, method, path, key)
	if !found {
		return fmt.Errorf("request %d %s %s has no header %s", index, method, path, key)
	}
	return expectEqual(value, t.substitute(expected), "header "+key)
}

func (t *testUtils) theRequestShouldHaveTheQueryEqualTo(index int, method, path, key, expected string) error {
	value, found := t.api.GetRequestQuery(index, method, path, key)
	if !found {
		return fmt.Errorf("request %d %s %s has no query parameter %s", index, method, path, key)
	}
	return expectEqual(value, t.substitute(expected), "query "+key)
}

func (t *testUtils) theRequestShouldHaveTheBodyEqualTo(index int, method, path string, expected *godog.DocString) error {
	body, found := t.api.GetRequestBody(index, method, path)
	if !found {
		return fmt.Errorf("request %d %s %s was not received", index, method, path)
	}
	return expectEqual(string(body), t.substitute(expected.Content), "body")
}

func (t *testUtils) theRequestShouldHaveBeenReceivedTimes(method, path string, times int) error {
	if count := t.api.GetCallCount(method, path); count != times {
		return fmt.Errorf("%s %s: expected %d calls, got %d", method, path, times, count)
	}
	return nil
}
