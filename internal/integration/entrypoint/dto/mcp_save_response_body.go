package dto

import (
	"path/filepath"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type MCPSaveResponseBody struct {
	Path string `json:"path"`
}

func (m *MCPSaveResponseBody) HasAbsolutePath() bool {
	return filepath.IsAbs(m.Path)
}

func (m *MCPSaveResponseBody) CleanPath() string {
	return filepath.Clean(m.Path)
}

func (m *MCPSaveResponseBody) SaveResponseBody(tabId string) SaveResponseBody {
	return NewSaveResponseBody(tabId, m.CleanPath())
}

func (m *MCPSaveResponseBody) WriteAIOutputFile() WriteAIOutputFile {
	return NewWriteAIOutputFile(m.CleanPath())
}

func (m *MCPSaveResponseBody) ActivityTarget() entity.AIActivityTarget {
	return entity.AIActivityTarget{Path: m.CleanPath()}
}
