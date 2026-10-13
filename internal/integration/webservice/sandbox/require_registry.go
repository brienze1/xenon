package sandbox

import (
	"embed"
	"encoding/json"
	"sync"

	"github.com/dop251/goja"
)

const (
	assetsDir     = "assets"
	manifestFile  = "MANIFEST.json"
	moduleWrapper = "(function (exports, require, module, global) {"
)

//go:embed assets/*.js assets/MANIFEST.json
var assetFiles embed.FS

type manifestEntry struct {
	Module  string `json:"module"`
	File    string `json:"file"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Sha256  string `json:"sha256"`
}

type manifest struct {
	Reference string          `json:"reference"`
	Files     []manifestEntry `json:"files"`
}

type moduleRegistry struct {
	files    map[string]string
	programs sync.Map
}

var (
	registryOnce     sync.Once
	registryInstance *moduleRegistry
)

func sharedModuleRegistry() *moduleRegistry {
	registryOnce.Do(func() {
		registryInstance = newModuleRegistry(readManifest())
	})
	return registryInstance
}

func readManifest() manifest {
	raw, err := assetFiles.ReadFile(assetsDir + "/" + manifestFile)
	if err != nil {
		panic(err)
	}
	var parsed manifest
	if err = json.Unmarshal(raw, &parsed); err != nil {
		panic(err)
	}
	return parsed
}

func newModuleRegistry(parsed manifest) *moduleRegistry {
	files := make(map[string]string, len(parsed.Files))
	for _, entry := range parsed.Files {
		files[entry.Module] = entry.File
	}
	return &moduleRegistry{files: files}
}

func (m *moduleRegistry) has(name string) bool {
	_, found := m.files[name]
	return found
}

func (m *moduleRegistry) program(name string) (*goja.Program, error) {
	if cached, found := m.programs.Load(name); found {
		return cached.(*goja.Program), nil
	}
	source, err := assetFiles.ReadFile(assetsDir + "/" + m.files[name])
	if err != nil {
		return nil, err
	}
	program, err := goja.Compile(name, moduleWrapper+string(source)+"\n})", false)
	if err != nil {
		return nil, err
	}
	m.programs.Store(name, program)
	return program, nil
}
