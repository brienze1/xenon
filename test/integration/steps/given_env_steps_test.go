package steps

import (
	"os"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenEnvSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the "([^"]*)" env var is set to "([^"]*)"$`, t.theEnvVarIsSetTo)
}

func (t *testUtils) theEnvVarIsSetTo(name, value string) error {
	return t.setEnvironment(name, t.substitute(value))
}

func (t *testUtils) setEnvironment(name, value string) error {
	if _, saved := t.environment[name]; !saved {
		t.environment[name] = os.Getenv(name)
	}
	return os.Setenv(name, value)
}
