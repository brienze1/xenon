package adapter

type AppProcess interface {
	Spawn(path string, env map[string]string) error
	Quit()
}
