package dto

type WriteAIOutputFile struct {
	Path string `json:"path" validate:"required,absolute_path"`
}

func NewWriteAIOutputFile(path string) WriteAIOutputFile {
	return WriteAIOutputFile{Path: path}
}

func (w *WriteAIOutputFile) FilePath() string {
	return w.Path
}
