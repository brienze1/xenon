package steps

import (
	"os"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenAppDataSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the app data file "([^"]*)" contains$`, t.theAppDataFileContains)
	ctx.Given(`^the app data file "([^"]*)" contains the text "([^"]*)"$`, t.theAppDataFileContainsTheText)
	ctx.Given(`^the app data path "([^"]*)" is removed$`, t.theAppDataPathIsRemoved)
}

func (t *testUtils) theAppDataFileContains(path string, content *godog.DocString) error {
	return writeFile(t.appDataPath(path), t.substitute(content.Content))
}

func (t *testUtils) theAppDataFileContainsTheText(path, text string) error {
	return writeFile(t.appDataPath(path), t.substitute(text))
}

func (t *testUtils) theAppDataPathIsRemoved(path string) error {
	return os.RemoveAll(t.appDataPath(path))
}
