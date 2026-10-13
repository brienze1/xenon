package steps

import (
	"net"
	"strconv"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenOAuthSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the "([^"]*)" variable is set to the mocked oauth provider url$`, t.theVariableIsSetToTheMockedOAuthProviderUrl)
	ctx.Given(`^the oauth provider is configured with$`, t.theOAuthProviderIsConfiguredWith)
	ctx.Given(`^the system browser ignores the pages it opens$`, t.theSystemBrowserIgnoresThePagesItOpens)
	ctx.Given(`^the "([^"]*)" variable is set to a free local port$`, t.theVariableIsSetToAFreeLocalPort)
	ctx.Step(`^the local port "([^"]*)" is in use$`, t.theLocalPortIsInUse)
}

func (t *testUtils) theVariableIsSetToTheMockedOAuthProviderUrl(name string) error {
	t.vars[name] = t.oauth.URL()
	return nil
}

func (t *testUtils) theOAuthProviderIsConfiguredWith(config *godog.DocString) error {
	return t.oauth.Configure(t.substitute(config.Content))
}

func (t *testUtils) theSystemBrowserIgnoresThePagesItOpens() error {
	t.browser.IgnoreOpenedPages()
	return nil
}

func (t *testUtils) theVariableIsSetToAFreeLocalPort(name string) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	t.vars[name] = strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	return listener.Close()
}

func (t *testUtils) theLocalPortIsInUse(port string) error {
	listener, err := net.Listen("tcp", "127.0.0.1:"+t.substitute(port))
	if err != nil {
		return err
	}
	t.busyPorts = append(t.busyPorts, listener)
	return nil
}
