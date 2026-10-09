package steps

import (
	"fmt"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerTrashSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the trash fails with "([^"]*)"$`, t.theTrashFailsWith)
	ctx.Then(`^the trash should contain (\d+) items$`, t.theTrashShouldContainItems)
	ctx.Then(`^the trash item (\d+) should be the workspace path "([^"]*)"$`, t.theTrashItemShouldBeTheWorkspacePath)
}

func (t *testUtils) theTrashFailsWith(message string) error {
	t.trash.Fail(message)
	return nil
}

func (t *testUtils) theTrashShouldContainItems(count int) error {
	if actual := t.trash.Count(); actual != count {
		return fmt.Errorf("trash: expected %d items, got %d", count, actual)
	}
	return nil
}

func (t *testUtils) theTrashItemShouldBeTheWorkspacePath(index int, path string) error {
	item, found := t.trash.Item(index)
	if !found {
		return fmt.Errorf("trash item %d does not exist", index)
	}
	return expectEqual(item, t.workspacePath(path), "trash item")
}
