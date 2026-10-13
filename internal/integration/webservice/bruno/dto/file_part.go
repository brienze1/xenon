package dto

type FilePart struct {
	Path        string
	ContentType string
	Selected    bool
}

func NewFilePart(path, contentType string, selected bool) FilePart {
	return FilePart{Path: path, ContentType: contentType, Selected: selected}
}
