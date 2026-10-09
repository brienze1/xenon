package steps

import (
	"fmt"
	"path/filepath"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenResponseSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the call should succeed$`, t.theCallShouldSucceed)
	ctx.Then(`^the call should fail with code "([^"]*)"$`, t.theCallShouldFailWithCode)
	ctx.Then(`^the error should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theErrorShouldContainTheFieldEqualTo)
	ctx.Then(`^the response should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theResponseShouldContainTheFieldEqualTo)
	ctx.Then(`^the response should contain the path field "([^"]*)" equal to "([^"]*)"$`, t.theResponseShouldContainThePathFieldEqualTo)
	ctx.Then(`^the response should contain the field "([^"]*)" array length equal to (\d+)$`, t.theResponseShouldContainTheFieldArrayLengthEqualTo)
	ctx.Then(`^the response field "([^"]*)" should exist$`, t.theResponseFieldShouldExist)
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
