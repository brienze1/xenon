package steps

import (
	"fmt"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenKeychainSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the keychain should contain an entry for the account "([^"]*)"$`, t.theKeychainShouldContainAnEntryForTheAccount)
	ctx.Then(`^the keychain should not contain an entry for the account "([^"]*)"$`, t.theKeychainShouldNotContainAnEntryForTheAccount)
	ctx.Then(`^the keychain should contain (\d+) entries$`, t.theKeychainShouldContainEntries)
}

func (t *testUtils) theKeychainShouldContainAnEntryForTheAccount(account string) error {
	if !t.keychain.HasAccount(t.substitute(account)) {
		return fmt.Errorf("the keychain has no entry for %q", t.substitute(account))
	}
	return nil
}

func (t *testUtils) theKeychainShouldNotContainAnEntryForTheAccount(account string) error {
	if t.keychain.HasAccount(t.substitute(account)) {
		return fmt.Errorf("the keychain still has an entry for %q", t.substitute(account))
	}
	return nil
}

func (t *testUtils) theKeychainShouldContainEntries(count int) error {
	if actual := t.keychain.Count(); actual != count {
		return fmt.Errorf("expected %d keychain entries, got %d", count, actual)
	}
	return nil
}
