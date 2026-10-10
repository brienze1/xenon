package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"

	"github.com/brienze1/xenon/pkg/logger"
)

func (t *testUtils) registerGivenVaultSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the keychain is available$`, t.theKeychainIsAvailable)
	ctx.Given(`^the keychain is unavailable$`, t.theKeychainIsUnavailable)
	ctx.Given(`^the keychain fails with "([^"]*)"$`, t.theKeychainFailsWith)
	ctx.Given(`^the keychain entry for the account "([^"]*)" is removed$`, t.theKeychainEntryForTheAccountIsRemoved)
	ctx.Given(`^the vault passphrase is "([^"]*)"$`, t.theVaultPassphraseIs)
	ctx.Step(`^the app is restarted$`, t.theAppIsRestarted)
	ctx.Given(`^the environment variable "([^"]*)" is set to "([^"]*)"$`, t.theEnvironmentVariableIsSetTo)
	ctx.Given(`^the log output is captured$`, t.theLogOutputIsCaptured)
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

func (t *testUtils) theVaultPassphraseIs(passphrase string) error {
	payload, _ := json.Marshal(map[string]string{"passphrase": passphrase, "confirmation": passphrase})
	if err := t.call(context.Background(), "VaultService", "SetupPassphrase", string(payload)); err != nil {
		return err
	}
	if t.callError != nil {
		return fmt.Errorf("setting up the vault passphrase failed: %s", t.callError)
	}
	return nil
}

func (t *testUtils) theAppIsRestarted() error {
	t.restarts++
	restarted := filepath.Join(t.tempRoot, fmt.Sprintf("app-data-restart-%d", t.restarts))
	if err := os.Rename(t.dataDir, restarted); err != nil {
		return err
	}
	t.dataDir = restarted
	t.vars["app_data"] = restarted
	return os.Setenv("XENON_DATA_DIR", restarted)
}

func (t *testUtils) theEnvironmentVariableIsSetTo(name, value string) error {
	return t.setEnvironment(name, t.substitute(value))
}

func (t *testUtils) theLogOutputIsCaptured() error {
	t.logs.Reset()
	logger.SetOutput(t.logs)
	return t.setEnvironment("LOG_LEVEL", "trace")
}

func (t *testUtils) setEnvironment(name, value string) error {
	if _, saved := t.environment[name]; !saved {
		t.environment[name] = os.Getenv(name)
	}
	return os.Setenv(name, value)
}
