package steps

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenApiBodySteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the body equal to$`, t.theRequestShouldHaveTheBodyEqualTo)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the body matching "([^"]*)"$`, t.theRequestShouldHaveTheBodyMatching)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the body field "([^"]*)" equal to "([^"]*)"$`, t.theRequestShouldHaveTheBodyFieldEqualTo)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the body field "([^"]*)" matching "([^"]*)"$`, t.theRequestShouldHaveTheBodyFieldMatching)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the body split by "([^"]*)" into (\d+) parts all matching "([^"]*)"$`, t.theRequestShouldHaveTheBodySplitIntoPartsAllMatching)
	ctx.Then(`^the (\d+) "([^"]*)" request to "([^"]*)" should have the body split by "([^"]*)" into (\d+) distinct parts$`, t.theRequestShouldHaveTheBodySplitIntoDistinctParts)
	ctx.Then(`^the (\d+) "([^"]*)" requests to "([^"]*)" should have different bodies$`, t.theRequestsShouldHaveDifferentBodies)
}

func (t *testUtils) theRequestShouldHaveTheBodyEqualTo(index int, method, path string, expected *godog.DocString) error {
	body, found := t.api.GetRequestBody(index, method, path)
	if !found {
		return fmt.Errorf("request %d %s %s was not received", index, method, path)
	}
	return expectEqual(string(body), t.substitute(expected.Content), "body")
}

func (t *testUtils) theRequestShouldHaveTheBodyMatching(index int, method, path, pattern string) error {
	body, found := t.api.GetRequestBody(index, method, path)
	if !found {
		return fmt.Errorf("request %d %s %s was not received", index, method, path)
	}
	return expectMatch(string(body), pattern, "body")
}

func (t *testUtils) theRequestShouldHaveTheBodyFieldEqualTo(index int, method, path, field, expected string) error {
	value, err := t.requestBodyField(index, method, path, field)
	if err != nil {
		return err
	}
	return expectEqual(stringify(value), t.substitute(expected), "body field "+field)
}

func (t *testUtils) theRequestShouldHaveTheBodyFieldMatching(index int, method, path, field, pattern string) error {
	value, err := t.requestBodyField(index, method, path, field)
	if err != nil {
		return err
	}
	return expectMatch(stringify(value), pattern, "body field "+field)
}

func (t *testUtils) theRequestShouldHaveTheBodySplitIntoPartsAllMatching(index int, method, path, separator string, count int, pattern string) error {
	parts, err := t.requestBodyParts(index, method, path, separator, count)
	if err != nil {
		return err
	}
	for position, part := range parts {
		if err := expectMatch(part, pattern, fmt.Sprintf("body part %d", position+1)); err != nil {
			return err
		}
	}
	return nil
}

func (t *testUtils) theRequestShouldHaveTheBodySplitIntoDistinctParts(index int, method, path, separator string, count int) error {
	parts, err := t.requestBodyParts(index, method, path, separator, count)
	if err != nil {
		return err
	}
	return expectDistinct(parts, "body parts")
}

func (t *testUtils) theRequestsShouldHaveDifferentBodies(count int, method, path string) error {
	bodies := make([]string, 0, count)
	for index := 1; index <= count; index++ {
		body, found := t.api.GetRequestBody(index, method, path)
		if !found {
			return fmt.Errorf("request %d %s %s was not received", index, method, path)
		}
		bodies = append(bodies, string(body))
	}
	return expectDistinct(bodies, "bodies")
}

func (t *testUtils) requestBodyField(index int, method, path, field string) (any, error) {
	body, found := t.api.GetRequestBody(index, method, path)
	if !found {
		return nil, fmt.Errorf("request %d %s %s was not received", index, method, path)
	}
	return fieldOf(body, field)
}

func (t *testUtils) requestBodyParts(index int, method, path, separator string, count int) ([]string, error) {
	body, found := t.api.GetRequestBody(index, method, path)
	if !found {
		return nil, fmt.Errorf("request %d %s %s was not received", index, method, path)
	}
	parts := strings.Split(string(body), separator)
	if len(parts) != count {
		return nil, fmt.Errorf("body of request %d %s %s has %d parts, expected %d", index, method, path, len(parts), count)
	}
	return parts, nil
}
