package dto

const (
	ImportFilesDroppedEvent = "import:files-dropped"
	ImportRequestedEvent    = "import:requested"
)

type ImportFilesDropped struct {
	Paths []string `json:"paths"`
}

func NewImportFilesDropped(paths []string) ImportFilesDropped {
	return ImportFilesDropped{Paths: append([]string{}, paths...)}
}

type ImportRequested struct {
	Source string `json:"source"`
}

func NewImportRequested(source string) ImportRequested {
	return ImportRequested{Source: source}
}
