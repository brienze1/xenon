package steps

import (
	"github.com/cucumber/godog"
)

func (t *testUtils) registerFileWatcherSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the file watcher reports a change to the workspace path "([^"]*)"$`, t.theFileWatcherReportsAChangeToTheWorkspacePath)
	ctx.Step(`^the file watcher reports the removal of the workspace path "([^"]*)"$`, t.theFileWatcherReportsTheRemovalOfTheWorkspacePath)
	ctx.Then(`^the file watcher should be watching the workspace path "([^"]*)"$`, t.theFileWatcherShouldBeWatchingTheWorkspacePath)
	ctx.Then(`^the file watcher should not be watching any folder$`, t.theFileWatcherShouldNotBeWatchingAnyFolder)
}

func (t *testUtils) theFileWatcherReportsAChangeToTheWorkspacePath(path string) error {
	t.watcher.Report(t.workspacePath(path), false)
	return nil
}

func (t *testUtils) theFileWatcherReportsTheRemovalOfTheWorkspacePath(path string) error {
	t.watcher.Report(t.workspacePath(path), true)
	return nil
}

func (t *testUtils) theFileWatcherShouldBeWatchingTheWorkspacePath(path string) error {
	return expectEqual(t.watcher.Root(), t.workspacePath(path), "watched folder")
}

func (t *testUtils) theFileWatcherShouldNotBeWatchingAnyFolder() error {
	return expectEqual(t.watcher.Root(), "", "watched folder")
}
