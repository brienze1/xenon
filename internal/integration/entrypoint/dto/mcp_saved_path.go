package dto

type MCPSavedPath struct {
	Path string `json:"path"`
}

func NewMCPSavedPath(path string) MCPSavedPath {
	return MCPSavedPath{Path: path}
}
