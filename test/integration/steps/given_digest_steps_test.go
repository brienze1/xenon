package steps

import "github.com/cucumber/godog"

func (t *testUtils) registerGivenDigestSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the "([^"]*)" variable is set to the mocked digest server url$`, t.theVariableIsSetToTheMockedDigestServerUrl)
	ctx.Given(`^the digest server is configured with$`, t.theDigestServerIsConfiguredWith)
}

func (t *testUtils) theVariableIsSetToTheMockedDigestServerUrl(name string) error {
	t.vars[name] = t.digest.URL()
	return nil
}

func (t *testUtils) theDigestServerIsConfiguredWith(config *godog.DocString) error {
	return t.digest.Configure(t.substitute(config.Content))
}
