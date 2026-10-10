package dto

type PickedSources struct {
	Paths []string `json:"paths"`
}

func NewPickedSources(paths []string) PickedSources {
	return PickedSources{Paths: append([]string{}, paths...)}
}
