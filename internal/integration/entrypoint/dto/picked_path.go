package dto

type PickedPath struct {
	Path string `json:"path"`
}

func NewPickedPath(path string) PickedPath {
	return PickedPath{Path: path}
}
