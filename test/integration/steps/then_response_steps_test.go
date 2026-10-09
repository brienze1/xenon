package steps

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenResponseSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the call should succeed$`, t.theCallShouldSucceed)
	ctx.Then(`^the call should fail with code "([^"]*)"$`, t.theCallShouldFailWithCode)
	ctx.Then(`^the error should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theErrorShouldContainTheFieldEqualTo)
	ctx.Then(`^the response should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theResponseShouldContainTheFieldEqualTo)
	ctx.Then(`^the response should contain the field "([^"]*)" equal to$`, t.theResponseShouldContainTheFieldEqualToTheText)
	ctx.Then(`^the response should contain the path field "([^"]*)" equal to "([^"]*)"$`, t.theResponseShouldContainThePathFieldEqualTo)
	ctx.Then(`^the response should contain the field "([^"]*)" array length equal to (\d+)$`, t.theResponseShouldContainTheFieldArrayLengthEqualTo)
	ctx.Then(`^the response should contain the field "([^"]*)" not equal to "([^"]*)"$`, t.theResponseShouldContainTheFieldNotEqualTo)
	ctx.Then(`^the response should contain the field "([^"]*)" containing "([^"]*)"$`, t.theResponseShouldContainTheFieldContaining)
	ctx.Then(`^the response should contain the field "([^"]*)" matching "([^"]*)"$`, t.theResponseShouldContainTheFieldMatching)
	ctx.Then(`^the response should contain the field "([^"]*)" with the items "([^"]*)"$`, t.theResponseShouldContainTheFieldWithTheItems)
	ctx.Then(`^the response field "([^"]*)" should exist$`, t.theResponseFieldShouldExist)
	ctx.Then(`^the response field "([^"]*)" should not exist$`, t.theResponseFieldShouldNotExist)
	ctx.Then(`^the response field "([^"]*)" should contain an item with "([^"]*)" equal to "([^"]*)" and "([^"]*)" equal to "([^"]*)"$`, t.theResponseFieldShouldContainAnItemWith)
	ctx.Then(`^the response field "([^"]*)" should contain (\d+) items with "([^"]*)" equal to "([^"]*)"$`, t.theResponseFieldShouldContainItemsWith)
}

func (t *testUtils) theCallShouldSucceed() error {
	if t.callError != nil {
		return fmt.Errorf("expected success, got error %s", t.callError)
	}
	return nil
}

func (t *testUtils) theCallShouldFailWithCode(code string) error {
	if t.callError == nil {
		return fmt.Errorf("expected error %s, got success %s", code, t.response)
	}
	actual, err := fieldOf(t.callError, "error.code")
	if err != nil {
		return err
	}
	return expectEqual(stringify(actual), code, "error code")
}

func (t *testUtils) theErrorShouldContainTheFieldEqualTo(path, expected string) error {
	if t.callError == nil {
		return fmt.Errorf("expected an error, got success %s", t.response)
	}
	actual, err := fieldOf(t.callError, path)
	if err != nil {
		return err
	}
	return expectEqual(stringify(actual), t.substitute(expected), path)
}

func (t *testUtils) theResponseShouldContainTheFieldEqualTo(path, expected string) error {
	actual, err := t.responseField(path)
	if err != nil {
		return err
	}
	return expectEqual(stringify(actual), t.substitute(expected), path)
}

func (t *testUtils) theResponseShouldContainTheFieldEqualToTheText(path string, expected *godog.DocString) error {
	return t.theResponseShouldContainTheFieldEqualTo(path, expected.Content)
}

func (t *testUtils) theResponseFieldShouldNotExist(path string) error {
	if err := t.theCallShouldSucceed(); err != nil {
		return err
	}
	if value, err := fieldOf(t.response, path); err == nil {
		return fmt.Errorf("field %q should not exist, got %v", path, value)
	}
	return nil
}

func (t *testUtils) theResponseFieldShouldContainAnItemWith(path, firstKey, firstValue, secondKey, secondValue string) error {
	items, err := t.responseItems(path)
	if err != nil {
		return err
	}
	expected := map[string]string{firstKey: t.substitute(firstValue), secondKey: t.substitute(secondValue)}
	for _, item := range items {
		if itemMatches(item, expected) {
			return nil
		}
	}
	return fmt.Errorf("field %q has no item with %v: %s", path, expected, stringify(items))
}

func (t *testUtils) theResponseFieldShouldContainItemsWith(path string, count int, key, value string) error {
	items, err := t.responseItems(path)
	if err != nil {
		return err
	}
	matches := 0
	for _, item := range items {
		if itemMatches(item, map[string]string{key: t.substitute(value)}) {
			matches++
		}
	}
	if matches != count {
		return fmt.Errorf("field %q: expected %d items with %s = %q, got %d: %s", path, count, key, value, matches, stringify(items))
	}
	return nil
}

func (t *testUtils) responseItems(path string) ([]any, error) {
	actual, err := t.responseField(path)
	if err != nil {
		return nil, err
	}
	items, isArray := actual.([]any)
	if !isArray {
		return nil, fmt.Errorf("field %q is not an array: %v", path, actual)
	}
	return items, nil
}

func itemMatches(item any, expected map[string]string) bool {
	for key, value := range expected {
		actual, found := lookup(item, key)
		if !found || stringify(actual) != value {
			return false
		}
	}
	return true
}

func (t *testUtils) theResponseShouldContainThePathFieldEqualTo(path, expected string) error {
	actual, err := t.responseField(path)
	if err != nil {
		return err
	}
	return expectEqual(filepath.Clean(stringify(actual)), filepath.Clean(filepath.FromSlash(t.substitute(expected))), path)
}

func (t *testUtils) theResponseShouldContainTheFieldArrayLengthEqualTo(path string, length int) error {
	actual, err := t.responseField(path)
	if err != nil {
		return err
	}
	items, isArray := actual.([]any)
	if !isArray {
		return fmt.Errorf("field %q is not an array: %v", path, actual)
	}
	if len(items) != length {
		return fmt.Errorf("field %q: expected %d items, got %d", path, length, len(items))
	}
	return nil
}

func (t *testUtils) theResponseFieldShouldExist(path string) error {
	_, err := t.responseField(path)
	return err
}

func (t *testUtils) responseField(path string) (any, error) {
	if err := t.theCallShouldSucceed(); err != nil {
		return nil, err
	}
	return fieldOf(t.response, path)
}

func (t *testUtils) theResponseShouldContainTheFieldNotEqualTo(path, unexpected string) error {
	actual, err := t.responseField(path)
	if err != nil {
		return err
	}
	if stringify(actual) == t.substitute(unexpected) {
		return fmt.Errorf("%s: expected anything but %q", path, stringify(actual))
	}
	return nil
}

func (t *testUtils) theResponseShouldContainTheFieldContaining(path, expected string) error {
	actual, err := t.responseField(path)
	if err != nil {
		return err
	}
	if !strings.Contains(stringify(actual), t.substitute(expected)) {
		return fmt.Errorf("%s: expected %q to contain %q", path, stringify(actual), expected)
	}
	return nil
}

func (t *testUtils) theResponseShouldContainTheFieldMatching(path, pattern string) error {
	actual, err := t.responseField(path)
	if err != nil {
		return err
	}
	if !regexp.MustCompile(pattern).MatchString(stringify(actual)) {
		return fmt.Errorf("%s: expected %q to match %s", path, stringify(actual), pattern)
	}
	return nil
}

func (t *testUtils) theResponseShouldContainTheFieldWithTheItems(path, items string) error {
	actual, err := t.responseField(path)
	if err != nil {
		return err
	}
	return expectItems(actual, t.substitute(items), path)
}

func expectItems(actual any, items, what string) error {
	values, isArray := actual.([]any)
	if !isArray {
		return fmt.Errorf("%s is not a list: %v", what, actual)
	}
	got := make([]string, 0, len(values))
	for _, value := range values {
		got = append(got, stringify(value))
	}
	return expectEqual(strings.Join(got, ", "), items, what)
}
