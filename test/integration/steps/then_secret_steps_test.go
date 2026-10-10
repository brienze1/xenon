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
