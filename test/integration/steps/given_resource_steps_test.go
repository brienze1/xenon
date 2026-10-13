package steps

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenResourceSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the resource folder "([^"]*)" is copied to the workspace folder "([^"]*)"$`, t.theResourceFolderIsCopiedToTheWorkspaceFolder)
}

func (t *testUtils) theResourceFolderIsCopiedToTheWorkspaceFolder(resource, folder string) error {
	source := resourcePath(resource)
	target := t.workspacePath(folder)
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		return os.WriteFile(destination, content, 0o644)
	})
}
