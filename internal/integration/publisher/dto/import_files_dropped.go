package dto

const ImportFilesDroppedEvent = "import:files-dropped"

type ImportFilesDropped struct {
	Paths []string `json:"paths"`
}

func NewImportFilesDropped(paths []string) ImportFilesDropped {
	return ImportFilesDropped{Paths: append([]string{}, paths...)}
}
