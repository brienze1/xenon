package steps

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/cucumber/godog"

	"github.com/brienze1/xenon/test/integration/mock"
)

type serviceShutdown interface {
	ServiceShutdown() error
}

func (t *testUtils) registerAppProcessSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the app shuts down$`, t.theAppShutsDown)
	ctx.Step(`^the app should have spawned (\d+) process(?:es)?$`, t.theAppShouldHaveSpawnedProcesses)
	ctx.Step(`^the (\d+) spawned process should run "([^"]*)"$`, t.theSpawnedProcessShouldRun)
	ctx.Step(`^the (\d+) spawned process should have the env var "([^"]*)" matching "([^"]*)"$`, t.theSpawnedProcessShouldHaveTheEnvVarMatching)
	ctx.Step(`^the app should have quit (\d+) times?$`, t.theAppShouldHaveQuit)
}

func (t *testUtils) theAppShutsDown() error {
	for _, service := range t.bindings {
		if stoppable, isStoppable := service.(serviceShutdown); isStoppable {
			if err := stoppable.ServiceShutdown(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *testUtils) theAppShouldHaveSpawnedProcesses(count int) error {
	return expectEqual(strconv.Itoa(len(t.process.Spawned())), strconv.Itoa(count), "spawned processes")
}

func (t *testUtils) theSpawnedProcessShouldRun(index int, path string) error {
	spawned, err := t.spawnedProcess(index)
	if err != nil {
		return err
	}
	return expectEqual(filepath.Clean(spawned.Path), filepath.Clean(filepath.FromSlash(t.substitute(path))), "spawned process path")
}

func (t *testUtils) theSpawnedProcessShouldHaveTheEnvVarMatching(index int, name, pattern string) error {
	spawned, err := t.spawnedProcess(index)
	if err != nil {
		return err
	}
	value, found := spawned.Env[name]
	if !found {
		return fmt.Errorf("spawned process %d has no env var %s", index, name)
	}
	return expectMatch(value, t.substitute(pattern), "env var "+name)
}

func (t *testUtils) theAppShouldHaveQuit(count int) error {
	return expectEqual(strconv.Itoa(t.process.Quits()), strconv.Itoa(count), "app quits")
}

func (t *testUtils) spawnedProcess(index int) (mock.SpawnedProcess, error) {
	spawned := t.process.Spawned()
	if index < 1 || index > len(spawned) {
		return mock.SpawnedProcess{}, fmt.Errorf("spawned process %d requested, %d spawned", index, len(spawned))
	}
	return spawned[index-1], nil
}
