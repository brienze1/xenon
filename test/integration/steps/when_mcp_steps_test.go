package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (t *testUtils) registerWhenMcpSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the MCP client is named "([^"]*)"$`, t.theMcpClientIsNamed)
	ctx.Step(`^I call the MCP tool "([^"]*)" with$`, t.iCallTheMcpToolWith)
	ctx.Step(`^I call the MCP tool "([^"]*)"$`, t.iCallTheMcpTool)
	ctx.Step(`^I call the MCP tool "([^"]*)" (\d+) times with$`, t.iCallTheMcpToolTimesWith)
	ctx.Step(`^I start calling the MCP tool "([^"]*)" in the background with$`, t.iStartCallingTheMcpToolInTheBackgroundWith)
	ctx.Step(`^I start calling the MCP tool "([^"]*)" in the background as "([^"]*)" with$`, t.iStartCallingTheMcpToolInTheBackgroundAsWith)
	ctx.Step(`^I list the MCP tools$`, t.iListTheMcpTools)
	ctx.Step(`^the MCP session is closed$`, t.theMcpSessionIsClosed)
}

func (t *testUtils) theMcpClientIsNamed(name string) error {
	t.mcpClientName = t.substitute(name)
	return nil
}

func (t *testUtils) theMcpSessionIsClosed() error {
	if session, open := t.mcpSessions[t.mcpClientName]; open {
		delete(t.mcpSessions, t.mcpClientName)
		return session.Close()
	}
	return nil
}

func (t *testUtils) iCallTheMcpToolWith(tool string, arguments *godog.DocString) error {
	return t.callMcpTool(tool, arguments.Content)
}

func (t *testUtils) iCallTheMcpTool(tool string) error {
	return t.callMcpTool(tool, "")
}

func (t *testUtils) iCallTheMcpToolTimesWith(tool string, times int, arguments *godog.DocString) error {
	for call := 0; call < times; call++ {
		if err := t.callMcpTool(tool, arguments.Content); err != nil {
			return err
		}
		if t.callError != nil {
			return fmt.Errorf("call %d of the MCP tool %s failed: %s", call+1, tool, t.callError)
		}
	}
	return nil
}

func (t *testUtils) callMcpTool(tool, arguments string) error {
	t.response, t.callError = nil, nil
	invoke, err := t.prepareMcpCall(tool, arguments)
	if err != nil {
		return err
	}
	result, err := invoke()
	t.response, t.callError, t.elapsed = result.response, result.callError, result.elapsed
	return err
}

func (t *testUtils) iStartCallingTheMcpToolInTheBackgroundWith(tool string, arguments *godog.DocString) error {
	invoke, err := t.prepareMcpCall(tool, arguments.Content)
	if err != nil {
		return err
	}
	call := &backgroundCall{done: make(chan struct{}), started: time.Now()}
	t.background = call
	go func() {
		call.result, call.err = invoke()
		close(call.done)
	}()
	return nil
}

func (t *testUtils) iStartCallingTheMcpToolInTheBackgroundAsWith(tool, name string, arguments *godog.DocString) error {
	if err := t.iStartCallingTheMcpToolInTheBackgroundWith(tool, arguments); err != nil {
		return err
	}
	t.backgrounds[name] = t.background
	return nil
}

func (t *testUtils) prepareMcpCall(tool, arguments string) (func() (callResult, error), error) {
	params := &mcp.CallToolParams{Name: tool, Arguments: map[string]any{}}
	if strings.TrimSpace(arguments) != "" {
		var decoded map[string]any
		if err := json.Unmarshal([]byte(t.substituteJson(arguments)), &decoded); err != nil {
			return nil, fmt.Errorf("invalid arguments for the MCP tool %s: %w", tool, err)
		}
		params.Arguments = decoded
	}
	connectCtx, cancel := context.WithTimeout(context.Background(), mcpCallTimeout)
	defer cancel()
	session, err := t.session(connectCtx)
	if err != nil {
		return nil, err
	}
	return func() (callResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), mcpCallTimeout)
		defer cancel()
		started := time.Now()
		result, err := session.CallTool(ctx, params)
		elapsed := time.Since(started)
		if err != nil {
			return callResult{elapsed: elapsed}, fmt.Errorf("calling the MCP tool %s: %w", tool, err)
		}
		text := toolText(result)
		if result.IsError {
			return callResult{callError: wrapToolError(text), elapsed: elapsed}, nil
		}
		return callResult{response: []byte(text), elapsed: elapsed}, nil
	}, nil
}

func (t *testUtils) iListTheMcpTools() error {
	t.response, t.callError = nil, nil
	ctx, cancel := context.WithTimeout(context.Background(), mcpCallTimeout)
	defer cancel()
	started := time.Now()
	session, err := t.session(ctx)
	if err != nil {
		return err
	}
	listed, err := session.ListTools(ctx, nil)
	t.elapsed = time.Since(started)
	if err != nil {
		return fmt.Errorf("listing the MCP tools: %w", err)
	}
	t.response, err = json.Marshal(listed.Tools)
	return err
}
