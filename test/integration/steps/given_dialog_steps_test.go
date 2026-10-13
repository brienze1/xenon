package steps

import (
	"path/filepath"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenDialogSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the (?:folder|file|save) (?:picker|dialog) will return "([^"]*)"$`, t.theFolderPickerWillReturn)
	ctx.Then(`^the (?:folder|file|save) (?:picker|dialog) should have been opened with the title "([^"]*)"$`, t.theFolderPickerShouldHaveBeenOpenedWithTheTitle)
	ctx.Then(`^the save dialog should have suggested the name "([^"]*)"$`, t.theSaveDialogShouldHaveSuggestedTheName)
}

func (t *testUtils) theFolderPickerWillReturn(path string) error {
	t.dialog.QueueAnswer(filepath.FromSlash(t.substitute(path)))
	return nil
}

func (t *testUtils) theFolderPickerShouldHaveBeenOpenedWithTheTitle(title string) error {
	return expectEqual(t.dialog.LastTitle(), title, "folder picker title")
}

func (t *testUtils) theSaveDialogShouldHaveSuggestedTheName(name string) error {
	return expectEqual(t.dialog.LastSuggestedName(), name, "suggested file name")
}
