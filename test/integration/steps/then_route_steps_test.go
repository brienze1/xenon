package steps

import (
	"errors"
	"fmt"
	"os"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenRouteSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the route response status should be (\d+)$`, t.theRouteResponseStatusShouldBe)
	ctx.Then(`^the route response header "([^"]*)" should be "([^"]*)"$`, t.theRouteResponseHeaderShouldBe)
	ctx.Then(`^the route response body should be$`, t.theRouteResponseBodyShouldBe)
	ctx.Then(`^the route response body should be the generated body of (\d+) bytes$`, t.theRouteResponseBodyShouldBeTheGeneratedBody)
	ctx.Then(`^the route response body should be the resource file "([^"]*)"$`, t.theRouteResponseBodyShouldBeTheResourceFile)
}

func (t *testUtils) theRouteResponseStatusShouldBe(status int) error {
	if t.routeResponse == nil {
		return errors.New("no route was requested")
	}
	if t.routeResponse.Code != status {
		return fmt.Errorf("route status: expected %d, got %d", status, t.routeResponse.Code)
	}
	return nil
}

func (t *testUtils) theRouteResponseHeaderShouldBe(key, expected string) error {
	if err := t.theRouteResponseStatusShouldBe(200); err != nil {
		return err
	}
	return expectEqual(t.routeResponse.Header().Get(key), t.substitute(expected), "route header "+key)
}

func (t *testUtils) theRouteResponseBodyShouldBe(expected *godog.DocString) error {
	return t.routeBodyShouldBe([]byte(t.substitute(expected.Content)))
}

func (t *testUtils) theRouteResponseBodyShouldBeTheGeneratedBody(size int) error {
	return t.routeBodyShouldBe(generatedBody(size))
}

func (t *testUtils) theRouteResponseBodyShouldBeTheResourceFile(name string) error {
	expected, err := os.ReadFile(resourcePath(name))
	if err != nil {
		return err
	}
	return t.routeBodyShouldBe(expected)
}

func (t *testUtils) routeBodyShouldBe(expected []byte) error {
	if err := t.theRouteResponseStatusShouldBe(200); err != nil {
		return err
	}
	return expectSameBytes(t.routeResponse.Body.Bytes(), expected, "route body")
}
