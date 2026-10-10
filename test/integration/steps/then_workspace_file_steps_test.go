package steps

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/cucumber/godog"
	"go.yaml.in/yaml/v3"
)

func (t *testUtils) registerThenWorkspaceFileSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the workspace file "([^"]*)" should exist$`, t.theWorkspaceFileShouldExist)
	ctx.Then(`^the workspace file "([^"]*)" should not exist$`, t.theWorkspaceFileShouldNotExist)
	ctx.Then(`^the workspace folder "([^"]*)" should exist$`, t.theWorkspaceFolderShouldExist)
	ctx.Then(`^the workspace file "([^"]*)" should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theWorkspaceFileShouldContainTheFieldEqualTo)
	ctx.Then(`^the workspace file "([^"]*)" should contain the field "([^"]*)" equal to$`, t.theWorkspaceFileShouldContainTheFieldEqualToTheText)
	ctx.Then(`^the workspace file "([^"]*)" should contain the field "([^"]*)" matching "([^"]*)"$`, t.theWorkspaceFileShouldContainTheFieldMatching)
	ctx.Then(`^the workspace file "([^"]*)" should contain the field "([^"]*)" array length equal to (\d+)$`, t.theWorkspaceFileShouldContainTheFieldArrayLengthEqualTo)
	ctx.Then(`^the workspace file "([^"]*)" should contain the field "([^"]*)" not equal to "([^"]*)"$`, t.theWorkspaceFileShouldContainTheFieldNotEqualTo)
	ctx.Then(`^the workspace file "([^"]*)" should contain the field "([^"]*)" with the items "([^"]*)"$`, t.theWorkspaceFileShouldContainTheFieldWithTheItems)
	ctx.Then(`^the workspace file "([^"]*)" should not contain the field "([^"]*)"$`, t.theWorkspaceFileShouldNotContainTheField)
	ctx.Then(`^the workspace file "([^"]*)" should equal$`, t.theWorkspaceFileShouldEqual)
	ctx.Then(`^the workspace file "([^"]*)" should be unchanged$`, t.theWorkspaceFileShouldBeUnchanged)
	ctx.Then(`^the workspace file "([^"]*)" should still have Windows line endings and a byte order mark$`, t.theWorkspaceFileShouldStillHaveWindowsLineEndingsAndAByteOrderMark)
	ctx.Then(`^the workspace folder "([^"]*)" should contain the entries "([^"]*)"$`, t.theWorkspaceFolderShouldContainTheEntries)
	ctx.Then(`^no file under the workspace folder should contain "([^"]*)"$`, t.noFileUnderTheWorkspaceFolderShouldContain)
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

func (t *testUtils) theWorkspaceFileShouldContainTheFieldEqualToTheText(path, field string, expected *godog.DocString) error {
	return t.theWorkspaceFileShouldContainTheFieldEqualTo(path, field, expected.Content)
}

func (t *testUtils) theWorkspaceFileShouldContainTheFieldMatching(path, field, pattern string) error {
	value, err := t.workspaceFileField(path, field)
	if err != nil {
		return err
	}
	return expectMatch(stringify(value), pattern, field)
}

func (t *testUtils) theWorkspaceFileShouldContainTheFieldArrayLengthEqualTo(path, field string, length int) error {
	value, err := t.workspaceFileField(path, field)
	if err != nil {
		return err
	}
	items, isArray := value.([]any)
	if !isArray {
		return fmt.Errorf("field %q of %s is not a list: %v", field, path, value)
	}
	if len(items) != length {
		return fmt.Errorf("field %q of %s: expected %d items, got %d", field, path, length, len(items))
	}
	return nil
}

func (t *testUtils) theWorkspaceFileShouldContainTheFieldNotEqualTo(path, field, unexpected string) error {
	value, err := t.workspaceFileField(path, field)
	if err != nil {
		return err
	}
	if stringify(value) == t.substitute(unexpected) {
		return fmt.Errorf("%s: expected anything but %q", field, stringify(value))
	}
	return nil
}

func (t *testUtils) theWorkspaceFileShouldContainTheFieldWithTheItems(path, field, items string) error {
	value, err := t.workspaceFileField(path, field)
	if err != nil {
		return err
	}
	return expectItems(value, t.substitute(items), field)
}

func (t *testUtils) theWorkspaceFileShouldNotContainTheField(path, field string) error {
	if value, err := t.workspaceFileField(path, field); err == nil {
		return fmt.Errorf("field %q of %s: expected no field, got %v", field, path, value)
	}
	return nil
}

func (t *testUtils) theWorkspaceFileShouldEqual(path string, expected *godog.DocString) error {
	content, err := os.ReadFile(t.workspacePath(path))
	if err != nil {
		return err
	}
	return expectEqual(string(content), t.substitute(expected.Content)+"\n", path)
}

func (t *testUtils) theWorkspaceFileShouldBeUnchanged(path string) error {
	expected, found := t.written[t.workspacePath(path)]
	if !found {
		return fmt.Errorf("%s was not written by a step of this scenario", path)
	}
	content, err := os.ReadFile(t.workspacePath(path))
	if err != nil {
		return err
	}
	return expectEqual(string(content), expected, path)
}

func (t *testUtils) theWorkspaceFileShouldStillHaveWindowsLineEndingsAndAByteOrderMark(path string) error {
	content, err := os.ReadFile(t.workspacePath(path))
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(content, byteOrderMark) || !bytes.Contains(content, []byte("\r\n")) {
		return fmt.Errorf("%s was rewritten: %q", path, content)
	}
	return nil
}

func (t *testUtils) theWorkspaceFolderShouldContainTheEntries(path, entries string) error {
	listing, err := os.ReadDir(t.workspacePath(path))
	if err != nil {
		return err
	}
	names := make([]string, 0, len(listing))
	for _, entry := range listing {
		names = append(names, entry.Name())
	}
	expected := strings.Split(t.substitute(entries), ", ")
	sort.Strings(names)
	sort.Strings(expected)
	return expectEqual(strings.Join(names, ", "), strings.Join(expected, ", "), path+" entries")
}

func (t *testUtils) noFileUnderTheWorkspaceFolderShouldContain(text string) error {
	return filesShouldNotContain(t.workspaceRoot, t.substitute(text), func(string) bool { return false })
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
