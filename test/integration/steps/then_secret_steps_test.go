package steps

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

const vaultFileExtension = ".age"

func (t *testUtils) registerThenSecretSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^no file under the workspace folder should contain "([^"]*)"$`, t.noFileUnderTheWorkspaceFolderShouldContain)
	ctx.Then(`^no file under the app data folder should contain "([^"]*)"$`, t.noFileUnderTheAppDataFolderShouldContain)
	ctx.Then(`^the keychain should contain an entry for the account "([^"]*)"$`, t.theKeychainShouldContainAnEntryForTheAccount)
	ctx.Then(`^the keychain should not contain an entry for the account "([^"]*)"$`, t.theKeychainShouldNotContainAnEntryForTheAccount)
	ctx.Then(`^the keychain should contain (\d+) entries$`, t.theKeychainShouldContainEntries)
	ctx.Then(`^the clipboard should contain "([^"]*)"$`, t.theClipboardShouldContain)
	ctx.Then(`^the log output should contain "([^"]*)"$`, t.theLogOutputShouldContain)
	ctx.Then(`^the log output should not contain "([^"]*)"$`, t.theLogOutputShouldNotContain)
	ctx.Then(`^the response should not contain "([^"]*)"$`, t.theResponseShouldNotContain)
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

func (t *testUtils) theClipboardShouldContain(text string) error {
	return expectEqual(t.clipboard.Text(), t.substitute(text), "clipboard")
}

func (t *testUtils) theLogOutputShouldContain(text string) error {
	if !t.logs.Contains(t.substitute(text)) {
		return fmt.Errorf("the log output does not contain %q", t.substitute(text))
	}
	return nil
}

func (t *testUtils) theLogOutputShouldNotContain(text string) error {
	if t.logs.Contains(t.substitute(text)) {
		return fmt.Errorf("the log output contains %q", t.substitute(text))
	}
	return nil
}

func (t *testUtils) theResponseShouldNotContain(text string) error {
	if bytes.Contains(t.response, []byte(t.substitute(text))) {
		return fmt.Errorf("the response contains %q: %s", t.substitute(text), t.response)
	}
	return nil
}

func (t *testUtils) noFileUnderTheWorkspaceFolderShouldContain(text string) error {
	return filesShouldNotContain(t.workspaceRoot, t.substitute(text), func(string) bool { return false })
}

func (t *testUtils) noFileUnderTheAppDataFolderShouldContain(text string) error {
	return filesShouldNotContain(t.dataDir, t.substitute(text), func(path string) bool {
		return strings.HasSuffix(path, vaultFileExtension)
	})
}

func filesShouldNotContain(root, text string, skip func(path string) bool) error {
	if text == "" {
		return fmt.Errorf("the searched text must not be empty")
	}
	matches := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || skip(path) {
			return err
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if bytes.Contains(content, []byte(text)) {
			relative, _ := filepath.Rel(root, path)
			matches = append(matches, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if len(matches) > 0 {
		return fmt.Errorf("%q was found in %s", text, strings.Join(matches, ", "))
	}
	return nil
}
