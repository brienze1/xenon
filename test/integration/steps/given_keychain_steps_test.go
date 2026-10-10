package steps

import (
	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenKeychainSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the keychain is available$`, t.theKeychainIsAvailable)
	ctx.Given(`^the keychain is unavailable$`, t.theKeychainIsUnavailable)
	ctx.Given(`^the keychain fails with "([^"]*)"$`, t.theKeychainFailsWith)
	ctx.Given(`^the keychain entry for the account "([^"]*)" is removed$`, t.theKeychainEntryForTheAccountIsRemoved)
}

func (t *testUtils) theKeychainIsAvailable() error {
	t.keychain.MakeAvailable()
	return nil
}

func (t *testUtils) theKeychainIsUnavailable() error {
	t.keychain.MakeUnavailable()
	return nil
}

func (t *testUtils) theKeychainFailsWith(message string) error {
	t.keychain.Fail(message)
	return nil
}

func (t *testUtils) theKeychainEntryForTheAccountIsRemoved(account string) error {
	t.keychain.RemoveAccount(t.substitute(account))
	return nil
}
