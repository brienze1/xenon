package mock

import "sync"

type SpawnedProcess struct {
	Env  map[string]string
	Path string
}

type AppProcess struct {
	spawned []SpawnedProcess
	quits   int
	mu      sync.RWMutex
}

func NewAppProcess() *AppProcess {
	return &AppProcess{}
}

func (a *AppProcess) Spawn(path string, env map[string]string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	copied := make(map[string]string, len(env))
	for key, value := range env {
		copied[key] = value
	}
	a.spawned = append(a.spawned, SpawnedProcess{Path: path, Env: copied})
	return nil
}

func (a *AppProcess) Quit() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.quits++
}

func (a *AppProcess) Spawned() []SpawnedProcess {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return append([]SpawnedProcess(nil), a.spawned...)
}

func (a *AppProcess) Quits() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.quits
}

func (a *AppProcess) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.spawned = nil
	a.quits = 0
}
