package steps

import (
	"net"
	"strings"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenVariableSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the "([^"]*)" variable is set to the mocked api url$`, t.theVariableIsSetToTheMockedApiUrl)
	ctx.Given(`^the "([^"]*)" variable is set to the mocked tls api url$`, t.theVariableIsSetToTheMockedTlsApiUrl)
	ctx.Given(`^the "([^"]*)" variable is set to a closed port url$`, t.theVariableIsSetToAClosedPortUrl)
	ctx.Given(`^the "([^"]*)" variable is set to "([^"]*)"$`, t.theVariableIsSetTo)
	ctx.Given(`^the "([^"]*)" variable is set to "([^"]*)" repeated (\d+) times separated by "([^"]*)"$`, t.theVariableIsSetToRepeatedTimesSeparatedBy)
	ctx.Step(`^I save the workspace file "([^"]*)" field "([^"]*)" as "([^"]*)"$`, t.iSaveTheWorkspaceFileFieldAs)
}

func (t *testUtils) theVariableIsSetToTheMockedApiUrl(name string) error {
	t.vars[name] = t.api.URL()
	return nil
}

func (t *testUtils) theVariableIsSetToTheMockedTlsApiUrl(name string) error {
	t.vars[name] = t.api.TLSURL()
	return nil
}

func (t *testUtils) theVariableIsSetToAClosedPortUrl(name string) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		return err
	}
	t.vars[name] = "http://" + address
	return nil
}

func (t *testUtils) theVariableIsSetTo(name, value string) error {
	t.vars[name] = t.substitute(value)
	return nil
}

func (t *testUtils) theVariableIsSetToRepeatedTimesSeparatedBy(name, value string, times int, separator string) error {
	parts := make([]string, times)
	for index := range parts {
		parts[index] = t.substitute(value)
	}
	t.vars[name] = strings.Join(parts, separator)
	return nil
}

func (t *testUtils) iSaveTheWorkspaceFileFieldAs(path, field, name string) error {
	value, err := t.workspaceFileField(path, field)
	if err != nil {
		return err
	}
	t.vars[name] = stringify(value)
	return nil
}
