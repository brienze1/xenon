package steps

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/cucumber/godog"
	"go.yaml.in/yaml/v3"
)

func (t *testUtils) registerThenFileSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the workspace file "([^"]*)" should exist$`, t.theWorkspaceFileShouldExist)
	ctx.Then(`^the workspace file "([^"]*)" should not exist$`, t.theWorkspaceFileShouldNotExist)
	ctx.Then(`^the workspace folder "([^"]*)" should exist$`, t.theWorkspaceFolderShouldExist)
	ctx.Then(`^the workspace file "([^"]*)" should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theWorkspaceFileShouldContainTheFieldEqualTo)
	ctx.Then(`^the workspace file "([^"]*)" should equal$`, t.theWorkspaceFileShouldEqual)
	ctx.Step(`^I save the workspace file "([^"]*)" field "([^"]*)" as "([^"]*)"$`, t.iSaveTheWorkspaceFileFieldAs)
	ctx.Then(`^the file "([^"]*)" should equal the resource file "([^"]*)"$`, t.theFileShouldEqualTheResourceFile)
	ctx.Then(`^the file "([^"]*)" should be the generated body of (\d+) bytes$`, t.theFileShouldBeTheGeneratedBody)
	ctx.Then(`^the app data file "([^"]*)" should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theAppDataFileShouldContainTheFieldEqualTo)
	ctx.Then(`^the app data file "([^"]*)" should contain the field "([^"]*)" array length equal to (\d+)$`, t.theAppDataFileShouldContainTheFieldArrayLengthEqualTo)
	ctx.Then(`^the app data file "([^"]*)" should not exist$`, t.theAppDataFileShouldNotExist)
	ctx.Then(`^the app data path "([^"]*)" should have the permissions "([^"]*)"$`, t.theAppDataPathShouldHaveThePermissions)
}

func (t *testUtils) theWorkspaceFileShouldExist(path string) error {
	info, err := os.Stat(t.workspacePath(path))
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a folder, expected a file", path)
	}
	return nil
}

func (t *testUtils) theWorkspaceFileShouldNotExist(path string) error {
	if _, err := os.Stat(t.workspacePath(path)); !os.IsNotExist(err) {
		return fmt.Errorf("expected %s not to exist (stat error: %v)", path, err)
	}
	return nil
}

func (t *testUtils) theWorkspaceFolderShouldExist(path string) error {
	info, err := os.Stat(t.workspacePath(path))
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is a file, expected a folder", path)
	}
	return nil
}

func (t *testUtils) theWorkspaceFileShouldContainTheFieldEqualTo(path, field, expected string) error {
	value, err := t.workspaceFileField(path, field)
	if err != nil {
		return err
	}
	return expectEqual(stringify(value), t.substitute(expected), field)
}

func (t *testUtils) theWorkspaceFileShouldEqual(path string, expected *godog.DocString) error {
	content, err := os.ReadFile(t.workspacePath(path))
	if err != nil {
		return err
	}
	return expectEqual(string(content), t.substitute(expected.Content)+"\n", path)
}

func (t *testUtils) iSaveTheWorkspaceFileFieldAs(path, field, name string) error {
	value, err := t.workspaceFileField(path, field)
	if err != nil {
		return err
	}
	t.vars[name] = stringify(value)
	return nil
}

func (t *testUtils) theFileShouldEqualTheResourceFile(path, name string) error {
	expected, err := os.ReadFile(resourcePath(name))
	if err != nil {
		return err
	}
	return t.fileShouldHaveBytes(path, expected)
}

func (t *testUtils) theFileShouldBeTheGeneratedBody(path string, size int) error {
	return t.fileShouldHaveBytes(path, generatedBody(size))
}

func (t *testUtils) fileShouldHaveBytes(path string, expected []byte) error {
	content, err := os.ReadFile(filepath.Join(t.tempRoot, filepath.FromSlash(t.substitute(path))))
	if err != nil {
		return err
	}
	return expectSameBytes(content, expected, path)
}

func expectSameBytes(actual, expected []byte, what string) error {
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("%s: expected %d bytes, got %d different bytes", what, len(expected), len(actual))
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

func (t *testUtils) theAppDataFileShouldNotExist(path string) error {
	if _, err := os.Stat(filepath.Join(t.dataDir, filepath.FromSlash(path))); !os.IsNotExist(err) {
		return fmt.Errorf("expected app data %s not to exist (stat error: %v)", path, err)
	}
	return nil
}

func (t *testUtils) theAppDataPathShouldHaveThePermissions(path, permissions string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	expected, err := strconv.ParseUint(permissions, 8, 32)
	if err != nil {
		return err
	}
	info, err := os.Stat(filepath.Join(t.dataDir, filepath.FromSlash(path)))
	if err != nil {
		return err
	}
	return expectEqual(fmt.Sprintf("%04o", info.Mode().Perm()), fmt.Sprintf("%04o", expected), path+" permissions")
}

func (t *testUtils) workspaceFileField(path, field string) (any, error) {
	content, err := os.ReadFile(t.workspacePath(path))
	if err != nil {
		return nil, err
	}
	var data any
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, err
	}
	value, found := lookup(data, field)
	if !found {
		return nil, fmt.Errorf("field %q not found in %s", field, path)
	}
	return value, nil
}

func (t *testUtils) appDataField(path, field string) (any, error) {
	content, err := os.ReadFile(filepath.Join(t.dataDir, filepath.FromSlash(path)))
	if err != nil {
		return nil, err
	}
	return fieldOf(content, field)
}
