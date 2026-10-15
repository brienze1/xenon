package adapter

import (
	"os"
	"os/exec"

	wails "github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon/internal/integration/adapter"
)

type osAppProcess struct{}

func NewAppProcess() adapter.AppProcess {
	return &osAppProcess{}
}

func (o *osAppProcess) Spawn(path string, env map[string]string) error {
	command := exec.Command(path)
	command.Env = os.Environ()
	for key, value := range env {
		command.Env = append(command.Env, key+"="+value)
	}
	detach(command)
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}

func (o *osAppProcess) Quit() {
	if app := wails.Get(); app != nil {
		app.Quit()
	}
}
