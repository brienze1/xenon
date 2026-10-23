package steps

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

const vaultFileExtension = ".age"

func (t *testUtils) registerThenAppDataSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the app data file "([^"]*)" should exist$`, t.theAppDataFileShouldExist)
	ctx.Then(`^the app data file "([^"]*)" should not exist$`, t.theAppDataFileShouldNotExist)
	ctx.Then(`^the app data file "([^"]*)" should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theAppDataFileShouldContainTheFieldEqualTo)
	ctx.Then(`^the app data file "([^"]*)" should contain the field "([^"]*)" equal to$`, t.theAppDataFileShouldContainTheFieldEqualToText)
	ctx.Then(`^the app data file "([^"]*)" should contain the field "([^"]*)" array length equal to (\d+)$`, t.theAppDataFileShouldContainTheFieldArrayLengthEqualTo)
	ctx.Then(`^the app data file "([^"]*)" should not contain "([^"]*)"$`, t.theAppDataFileShouldNotContain)
	ctx.Then(`^the app data file "([^"]*)" should contain "([^"]*)"$`, t.theAppDataFileShouldContain)
	ctx.Then(`^the app data file "([^"]*)" should have (\d+) lines$`, t.theAppDataFileShouldHaveLines)
	ctx.Then(`^the app data file "([^"]*)" line (\d+) should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theAppDataFileLineShouldContainTheFieldEqualTo)
	ctx.Then(`^the app data path "([^"]*)" should have the permissions "([^"]*)"$`, t.theAppDataPathShouldHaveThePermissions)
	ctx.Then(`^no file under the app data folder should contain "([^"]*)"$`, t.noFileUnderTheAppDataFolderShouldContain)
}

func (t *testUtils) theAppDataFileShouldExist(path string) error {
	info, err := os.Stat(t.appDataPath(path))
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("app data %s is a folder, expected a file", path)
	}
	return nil
}

func (t *testUtils) theAppDataFileShouldNotExist(path string) error {
	if _, err := os.Stat(t.appDataPath(path)); !os.IsNotExist(err) {
		return fmt.Errorf("expected app data %s not to exist (stat error: %v)", path, err)
	}
	return nil
}

func (t *testUtils) theAppDataFileShouldContainTheFieldEqualTo(path, field, expected string) error {
	value, err := t.appDataField(path, field)
	if err != nil {
		return err
	}
	return expectEqual(stringify(value), t.substitute(expected), field)
}

func (t *testUtils) theAppDataFileShouldContainTheFieldEqualToText(path, field string, expected *godog.DocString) error {
	return t.theAppDataFileShouldContainTheFieldEqualTo(path, field, expected.Content)
}

func (t *testUtils) theAppDataFileShouldContainTheFieldArrayLengthEqualTo(path, field string, length int) error {
	value, err := t.appDataField(path, field)
	if err != nil {
		return err
	}
	items, isArray := value.([]any)
	if !isArray || len(items) != length {
		return fmt.Errorf("field %q: expected %d items, got %v", field, length, value)
	}
	return nil
}

func (t *testUtils) theAppDataFileShouldNotContain(path, text string) error {
	content, err := os.ReadFile(t.appDataPath(path))
	if err != nil {
		return err
	}
	if bytes.Contains(content, []byte(t.substitute(text))) {
		return fmt.Errorf("app data %s contains %q", path, t.substitute(text))
	}
	return nil
}

func (t *testUtils) theAppDataFileShouldContain(path, text string) error {
	content, err := os.ReadFile(t.appDataPath(path))
	if err != nil {
		return err
	}
	if !bytes.Contains(content, []byte(t.substitute(text))) {
		return fmt.Errorf("app data %s does not contain %q", path, t.substitute(text))
	}
	return nil
}

func (t *testUtils) theAppDataFileShouldHaveLines(path string, count int) error {
	lines, err := t.appDataLines(path)
	if err != nil {
		return err
	}
	if len(lines) != count {
		return fmt.Errorf("app data %s: expected %d lines, got %d", path, count, len(lines))
	}
	return nil
}

func (t *testUtils) theAppDataFileLineShouldContainTheFieldEqualTo(path string, line int, field, expected string) error {
	lines, err := t.appDataLines(path)
	if err != nil {
		return err
	}
	if line < 1 || line > len(lines) {
		return fmt.Errorf("app data %s has %d lines, no line %d", path, len(lines), line)
	}
	value, err := fieldOf([]byte(lines[line-1]), field)
	if err != nil {
		return err
	}
	return expectEqual(stringify(value), t.substitute(expected), field)
}

func (t *testUtils) theAppDataPathShouldHaveThePermissions(path, permissions string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	expected, err := strconv.ParseUint(permissions, 8, 32)
	if err != nil {
		return err
	}
	info, err := os.Stat(t.appDataPath(path))
	if err != nil {
		return err
	}
	return expectEqual(fmt.Sprintf("%04o", info.Mode().Perm()), fmt.Sprintf("%04o", expected), path+" permissions")
}

func (t *testUtils) noFileUnderTheAppDataFolderShouldContain(text string) error {
	return filesShouldNotContain(t.dataDir, t.substitute(text), func(path string) bool {
		return strings.HasSuffix(path, vaultFileExtension)
	})
}

func (t *testUtils) appDataField(path, field string) (any, error) {
	content, err := os.ReadFile(t.appDataPath(path))
	if err != nil {
		return nil, err
	}
	return fieldOf(content, t.substitute(field))
}

func (t *testUtils) appDataLines(path string) ([]string, error) {
	content, err := os.ReadFile(t.appDataPath(path))
	if err != nil {
		return nil, err
	}
	lines := []string{}
	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}
