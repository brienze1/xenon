package mcpbridge

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	appFileName = "app.json"
	defaultPort = 47821
)

type appFile struct {
	Settings appSettings `json:"settings"`
}

type appSettings struct {
	AIAccess aiAccess `json:"aiAccess"`
}

type aiAccess struct {
	Enabled *bool `json:"enabled"`
	Port    int   `json:"port"`
}

// readAIAccess never fails: a missing or unreadable app.json reads as the default settings.
func readAIAccess(dataDir string) aiAccess {
	content, err := os.ReadFile(filepath.Join(dataDir, appFileName))
	if err != nil {
		return aiAccess{}
	}
	var file appFile
	if err := json.Unmarshal(content, &file); err != nil {
		return aiAccess{}
	}
	return file.Settings.AIAccess
}

func (a aiAccess) isEnabled() bool {
	return a.Enabled == nil || *a.Enabled
}

// resolvedPort prefers the saved port; the override only replaces the default.
func (a aiAccess) resolvedPort(override *int) int {
	if a.Port > 0 {
		return a.Port
	}
	if override != nil {
		return *override
	}
	return defaultPort
}
