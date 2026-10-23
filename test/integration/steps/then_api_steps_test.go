package steps

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenApiSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the "([^"]*)" request to "([^"]*)" should have been received (\d+) times$`, t.theRequestShouldHaveBeenReceivedTimes)
	ctx.Then(`^the "([^"]*)" request to "([^"]*)" should have been received between (\d+) and (\d+) times$`, t.theRequestShouldHaveBeenReceivedBetweenTimes)
	ctx.Then(`^the mock API should have received the requests "([^"]*)" in order$`, t.theMockAPIShouldHaveReceivedTheRequestsInOrder)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have been received at least (\d+) ms after the (\d+) "([^"]*)" request to "([^"]*)"$`, t.theRequestShouldHaveBeenReceivedAtLeastAfter)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the header "([^"]*)" equal to "([^"]*)"$`, t.theRequestShouldHaveTheHeaderEqualTo)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the header "([^"]*)" matching "([^"]*)"$`, t.theRequestShouldHaveTheHeaderMatching)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should not have the header "([^"]*)"$`, t.theRequestShouldNotHaveTheHeader)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the header "([^"]*)" equal to the resource file "([^"]*)"$`, t.theRequestShouldHaveTheHeaderEqualToTheResourceFile)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the query "([^"]*)" equal to "([^"]*)"$`, t.theRequestShouldHaveTheQueryEqualTo)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the query "([^"]*)" matching "([^"]*)"$`, t.theRequestShouldHaveTheQueryMatching)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the raw query "([^"]*)"$`, t.theRequestShouldHaveTheRawQuery)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the form field "([^"]*)" equal to "([^"]*)"$`, t.theRequestShouldHaveTheFormFieldEqualTo)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should not have the form field "([^"]*)"$`, t.theRequestShouldNotHaveTheFormField)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the form file "([^"]*)" named "([^"]*)" with content "([^"]*)"$`, t.theRequestShouldHaveTheFormFileNamedWithContent)
}

func (t *testUtils) theRequestShouldHaveBeenReceivedTimes(method, path string, times int) error {
	if count := t.api.GetCallCount(method, path); count != times {
		return fmt.Errorf("%s %s: expected %d calls, got %d", method, path, times, count)
	}
	return nil
}

func (t *testUtils) theRequestShouldHaveBeenReceivedBetweenTimes(method, path string, minimum, maximum int) error {
	if count := t.api.GetCallCount(method, path); count < minimum || count > maximum {
		return fmt.Errorf("%s %s: expected between %d and %d calls, got %d", method, path, minimum, maximum, count)
	}
	return nil
}

func (t *testUtils) theMockAPIShouldHaveReceivedTheRequestsInOrder(expected string) error {
	return expectEqual(strings.Join(t.api.Sequence(), ", "), t.substitute(expected), "request order")
}

func (t *testUtils) theRequestShouldHaveBeenReceivedAtLeastAfter(index int, method, path string, milliseconds, previousIndex int, previousMethod, previousPath string) error {
	at, found := t.api.GetReceivedAt(index, method, path)
	if !found {
		return fmt.Errorf("request %d %s %s was not received", index, method, path)
	}
	previous, found := t.api.GetReceivedAt(previousIndex, previousMethod, previousPath)
	if !found {
		return fmt.Errorf("request %d %s %s was not received", previousIndex, previousMethod, previousPath)
	}
	if gap := at.Sub(previous); gap < time.Duration(milliseconds)*time.Millisecond {
		return fmt.Errorf("request %d %s %s came %s after request %d %s %s, expected at least %d ms", index, method, path, gap, previousIndex, previousMethod, previousPath, milliseconds)
	}
	return nil
}

func (t *testUtils) theRequestShouldHaveTheHeaderEqualTo(index int, method, path, key, expected string) error {
	value, found := t.api.GetRequestHeader(index, method, path, key)
	if !found {
		return fmt.Errorf("request %d %s %s has no header %s", index, method, path, key)
	}
	return expectEqual(value, t.substitute(expected), "header "+key)
}

func (t *testUtils) theRequestShouldHaveTheHeaderMatching(index int, method, path, key, pattern string) error {
	value, found := t.api.GetRequestHeader(index, method, path, key)
	if !found {
		return fmt.Errorf("request %d %s %s has no header %s", index, method, path, key)
	}
	return expectMatch(value, pattern, "header "+key)
}

func (t *testUtils) theRequestShouldNotHaveTheHeader(index int, method, path, key string) error {
	if value, found := t.api.GetRequestHeader(index, method, path, key); found {
		return fmt.Errorf("request %d %s %s has the header %s = %q", index, method, path, key, value)
	}
	return nil
}

func (t *testUtils) theRequestShouldHaveTheQueryEqualTo(index int, method, path, key, expected string) error {
	value, found := t.api.GetRequestQuery(index, method, path, key)
	if !found {
		return fmt.Errorf("request %d %s %s has no query parameter %s", index, method, path, key)
	}
	return expectEqual(value, t.substitute(expected), "query "+key)
}

func (t *testUtils) theRequestShouldHaveTheQueryMatching(index int, method, path, key, pattern string) error {
	value, found := t.api.GetRequestQuery(index, method, path, key)
	if !found {
		return fmt.Errorf("request %d %s %s has no query parameter %s", index, method, path, key)
	}
	return expectMatch(value, pattern, "query "+key)
}

func (t *testUtils) theRequestShouldHaveTheRawQuery(index int, method, path, expected string) error {
	value, found := t.api.GetRequestRawQuery(index, method, path)
	if !found {
		return fmt.Errorf("request %d %s %s was not received", index, method, path)
	}
	return expectEqual(value, t.substitute(expected), "raw query")
}

func (t *testUtils) theRequestShouldHaveTheFormFieldEqualTo(index int, method, path, key, expected string) error {
	value, found := t.api.GetRequestFormValue(index, method, path, key)
	if !found {
		return fmt.Errorf("request %d %s %s has no form field %s", index, method, path, key)
	}
	return expectEqual(value, t.substitute(expected), "form field "+key)
}

func (t *testUtils) theRequestShouldNotHaveTheFormField(index int, method, path, key string) error {
	if value, found := t.api.GetRequestFormValue(index, method, path, key); found {
		return fmt.Errorf("request %d %s %s has the form field %s = %q", index, method, path, key, value)
	}
	return nil
}

func (t *testUtils) theRequestShouldHaveTheFormFileNamedWithContent(index int, method, path, key, name, content string) error {
	fileName, data, found := t.api.GetRequestFormFile(index, method, path, key)
	if !found {
		return fmt.Errorf("request %d %s %s has no form file %s", index, method, path, key)
	}
	if err := expectEqual(fileName, name, "form file name"); err != nil {
		return err
	}
	return expectEqual(string(data), t.substitute(content), "form file content")
}

func (t *testUtils) theRequestShouldHaveTheHeaderEqualToTheResourceFile(index int, method, path, key, resource string) error {
	expected, err := os.ReadFile(resourcePath(resource))
	if err != nil {
		return err
	}
	return t.theRequestShouldHaveTheHeaderEqualTo(index, method, path, key, strings.TrimSpace(string(expected)))
}
