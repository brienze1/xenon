package steps

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerFileManagerSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the file manager fails with "([^"]*)"$`, t.theFileManagerFailsWith)
	ctx.Then(`^the file manager should have shown "([^"]*)"$`, t.theFileManagerShouldHaveShown)
}

func (t *testUtils) theFileManagerFailsWith(message string) error {
	t.fileManager.Fail(message)
	return nil
}

func (t *testUtils) theFileManagerShouldHaveShown(path string) error {
	expected := filepath.FromSlash(t.substitute(path))
	for _, shown := range t.fileManager.Shown() {
		if shown == expected {
			return nil
		}
	}
	return fmt.Errorf("the file manager showed [%s], expected %s", strings.Join(t.fileManager.Shown(), ", "), expected)
}
