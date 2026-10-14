package steps

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenResponseMatchSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the response should contain the field "([^"]*)" matching "([^"]*)"$`, t.theResponseShouldContainTheFieldMatching)
	ctx.Then(`^the response should match the field "([^"]*)" of the resource file "([^"]*)"$`, t.theResponseShouldMatchTheFieldOfTheResourceFile)
	ctx.Then(`^the response field "([^"]*)" should equal the resource file "([^"]*)"$`, t.theResponseFieldShouldEqualTheResourceFile)
}

func (t *testUtils) theResponseShouldContainTheFieldMatching(path, pattern string) error {
	actual, err := t.responseField(path)
	if err != nil {
		return err
	}
	return expectMatch(stringify(actual), strings.ReplaceAll(pattern, `\\`, `\`), path)
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

func (t *testUtils) theResponseFieldShouldEqualTheResourceFile(path, name string) error {
	actual, err := t.responseField(path)
	if err != nil {
		return err
	}
	expected, err := os.ReadFile(resourcePath(t.substitute(name)))
	if err != nil {
		return err
	}
	if stringify(actual) != string(expected) {
		return fmt.Errorf("%s differs from the resource %s:\n--- expected ---\n%s\n--- actual ---\n%s", path, name, expected, stringify(actual))
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
