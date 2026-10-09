package steps

import (
	"path/filepath"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenDialogSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the folder picker will return "([^"]*)"$`, t.theFolderPickerWillReturn)
	ctx.Then(`^the folder picker should have been opened with the title "([^"]*)"$`, t.theFolderPickerShouldHaveBeenOpenedWithTheTitle)
}

func (t *testUtils) theFolderPickerWillReturn(path string) error {
	t.dialog.QueueAnswer(filepath.FromSlash(t.substitute(path)))
	return nil
}

func (t *testUtils) theFolderPickerShouldHaveBeenOpenedWithTheTitle(title string) error {
	return expectEqual(t.dialog.LastTitle(), title, "folder picker title")
}
