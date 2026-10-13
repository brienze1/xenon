package steps

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenResponseListSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the response should contain the field "([^"]*)" array length equal to (\d+)$`, t.theResponseShouldContainTheFieldArrayLengthEqualTo)
	ctx.Then(`^the response should contain the field "([^"]*)" with the items "([^"]*)"$`, t.theResponseShouldContainTheFieldWithTheItems)
	ctx.Then(`^the response field "([^"]*)" should contain an item with "([^"]*)" equal to "([^"]*)" and "([^"]*)" equal to "([^"]*)"$`, t.theResponseFieldShouldContainAnItemWith)
	ctx.Then(`^the response field "([^"]*)" should contain (\d+) items with "([^"]*)" equal to "([^"]*)"$`, t.theResponseFieldShouldContainItemsWith)
	ctx.Then(`^every item of the saved "([^"]*)" should have the field "([^"]*)" in "([^"]*)"$`, t.everyItemOfTheSavedShouldHaveTheFieldIn)
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

func (t *testUtils) everyItemOfTheSavedShouldHaveTheFieldIn(saved, field, allowed string) error {
	items, err := decodeJson([]byte(t.vars[saved]))
	if err != nil {
		return err
	}
	list, isArray := items.([]any)
	if !isArray || len(list) == 0 {
		return fmt.Errorf("saved %q is not a non-empty list: %v", saved, items)
	}
	accepted := strings.Split(t.vars[allowed], ",")
	for index, item := range list {
		value, found := lookup(item, field)
		if !found || !slices.Contains(accepted, stringify(value)) {
			return fmt.Errorf("item %d of %q has %s = %v, not one of %v", index, saved, field, value, accepted)
		}
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

func (t *testUtils) theResponseShouldContainTheFieldArrayLengthEqualTo(path string, length int) error {
	actual, err := t.responseField(path)
	if err != nil {
		return err
	}
	return expectLength(actual, length, "field "+strconv.Quote(path))
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
