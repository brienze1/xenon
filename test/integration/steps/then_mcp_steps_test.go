package steps

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenMcpSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the error should not contain "([^"]*)"$`, t.theErrorShouldNotContain)
	ctx.Then(`^the response should have at most (\d+) characters$`, t.theResponseShouldHaveAtMostCharacters)
	ctx.Then(`^the response field "([^"]*)" should have (\d+) characters$`, t.theResponseFieldShouldHaveCharacters)
	ctx.Then(`^every item of the response field "([^"]*)" should have the field "([^"]*)" matching "([^"]*)"$`, t.everyItemOfTheResponseFieldShouldHaveTheFieldMatching)
	ctx.Then(`^the Xenon window should have been brought to the front (\d+) times?$`, t.theXenonWindowShouldHaveBeenBroughtToTheFrontTimes)
	ctx.Then(`^the background call "([^"]*)" should finish within (\d+) ms$`, t.theNamedBackgroundCallShouldFinishWithin)
	ctx.Then(`^the MCP endpoint should refuse connections$`, t.theMcpEndpointShouldRefuseConnections)
}

func (t *testUtils) theErrorShouldNotContain(text string) error {
	if bytes.Contains(t.callError, []byte(t.substitute(text))) {
		return fmt.Errorf("the error contains %q: %s", t.substitute(text), t.callError)
	}
	return nil
}

func (t *testUtils) theResponseShouldHaveAtMostCharacters(limit int) error {
	if size := len([]rune(string(t.response))); size > limit {
		return fmt.Errorf("the response has %d characters, expected at most %d", size, limit)
	}
	return nil
}

func (t *testUtils) theXenonWindowShouldHaveBeenBroughtToTheFrontTimes(times int) error {
	if actual := t.window.Count(); actual != times {
		return fmt.Errorf("the Xenon window was brought to the front %d times, expected %d", actual, times)
	}
	return nil
}

func (t *testUtils) theResponseFieldShouldHaveCharacters(path string, length int) error {
	value, err := t.responseField(path)
	if err != nil {
		return err
	}
	if actual := len([]rune(stringify(value))); actual != length {
		return fmt.Errorf("%s: expected %d characters, got %d", path, length, actual)
	}
	return nil
}

func (t *testUtils) everyItemOfTheResponseFieldShouldHaveTheFieldMatching(path, key, pattern string) error {
	items, err := t.responseItems(path)
	if err != nil {
		return err
	}
	expression, err := regexp.Compile(t.substitute(pattern))
	if err != nil {
		return err
	}
	for _, item := range items {
		value, found := lookup(item, key)
		if !found || !expression.MatchString(stringify(value)) {
			return fmt.Errorf("%s: an item has %s = %q, expected it to match %q: %s", path, key, stringify(value), pattern, stringify(item))
		}
	}
	return nil
}

func (t *testUtils) theNamedBackgroundCallShouldFinishWithin(name string, milliseconds int) error {
	call, found := t.backgrounds[name]
	if !found {
		return fmt.Errorf("no background call named %q was started", name)
	}
	t.background = call
	return t.theBackgroundCallShouldFinishWithin(milliseconds)
}

func (t *testUtils) theMcpEndpointShouldRefuseConnections() error {
	connection, err := net.DialTimeout("tcp", "127.0.0.1:"+t.vars["mcp_port"], time.Second)
	if err == nil {
		_ = connection.Close()
		return fmt.Errorf("the MCP endpoint %s accepted a connection", t.mcpEndpoint())
	}
	return nil
}
