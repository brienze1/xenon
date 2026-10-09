package dto

type FileChange struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Path    string `json:"path" validate:"required,absolute_path"`
}

func NewFileChange(localId, path string) FileChange {
	return FileChange{LocalId: localId, Path: path}
}
