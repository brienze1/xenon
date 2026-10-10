package steps

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenResponseSteps(ctx *godog.ScenarioContext) {
	t.registerThenResponseListSteps(ctx)
	ctx.Then(`^the call should succeed$`, t.theCallShouldSucceed)
	ctx.Then(`^the last call should have taken less than (\d+) ms$`, t.theLastCallShouldHaveTakenLessThan)
	ctx.Then(`^the call should fail with code "([^"]*)"$`, t.theCallShouldFailWithCode)
	ctx.Then(`^the error should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theErrorShouldContainTheFieldEqualTo)
	ctx.Then(`^the response should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theResponseShouldContainTheFieldEqualTo)
	ctx.Then(`^the response should contain the field "([^"]*)" equal to$`, t.theResponseShouldContainTheFieldEqualToTheText)
	ctx.Then(`^the response should contain the path field "([^"]*)" equal to "([^"]*)"$`, t.theResponseShouldContainThePathFieldEqualTo)
	ctx.Then(`^the response should contain the field "([^"]*)" not equal to "([^"]*)"$`, t.theResponseShouldContainTheFieldNotEqualTo)
	ctx.Then(`^the response should contain the field "([^"]*)" containing "([^"]*)"$`, t.theResponseShouldContainTheFieldContaining)
	ctx.Then(`^the response should contain the field "([^"]*)" matching "([^"]*)"$`, t.theResponseShouldContainTheFieldMatching)
	ctx.Then(`^the response field "([^"]*)" should exist$`, t.theResponseFieldShouldExist)
	ctx.Then(`^the response field "([^"]*)" should not exist$`, t.theResponseFieldShouldNotExist)
	ctx.Then(`^the response should match the field "([^"]*)" of the resource file "([^"]*)"$`, t.theResponseShouldMatchTheFieldOfTheResourceFile)
}

func (t *testUtils) theCallShouldSucceed() error {
	if t.callError != nil {
		return fmt.Errorf("expected success, got error %s", t.callError)
	}
	return nil
}

func (t *testUtils) theLastCallShouldHaveTakenLessThan(milliseconds int) error {
	if limit := time.Duration(milliseconds) * time.Millisecond; t.elapsed >= limit {
		return fmt.Errorf("the last call took %s, expected less than %s", t.elapsed, limit)
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

func (t *testUtils) theResponseShouldContainThePathFieldEqualTo(path, expected string) error {
	actual, err := t.responseField(path)
	if err != nil {
		return err
	}
	return expectEqual(filepath.Clean(stringify(actual)), filepath.Clean(filepath.FromSlash(t.substitute(expected))), path)
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
	return expectMatch(stringify(actual), pattern, path)
}

func (t *testUtils) theResponseShouldMatchTheFieldOfTheResourceFile(field, resource string) error {
	expected, err := resourceFileField(resource, field)
	if err != nil {
		return err
	}
	actual, err := decodeJson(t.response)
	if err != nil {
		return err
	}
	if err := matchFields(actual, expected, ""); err != nil {
		return fmt.Errorf("%w in %s", err, t.response)
	}
	return nil
}

func matchFields(actual, expected any, path string) error {
	switch want := expected.(type) {
	case map[string]any:
		for key, value := range want {
			child, found := lookup(actual, key)
			if !found {
				return fmt.Errorf("field %q is missing", joinFieldPath(path, key))
			}
			if err := matchFields(child, value, joinFieldPath(path, key)); err != nil {
				return err
			}
		}
		return nil
	case []any:
		got, isList := actual.([]any)
		if !isList || len(got) != len(want) {
			return fmt.Errorf("field %q: expected %d items, got %s", path, len(want), stringify(actual))
		}
		for index := range want {
			if err := matchFields(got[index], want[index], joinFieldPath(path, strconv.Itoa(index))); err != nil {
				return err
			}
		}
		return nil
	default:
		return expectEqual(stringify(actual), stringify(want), "field "+strconv.Quote(path))
	}
}

func joinFieldPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}
