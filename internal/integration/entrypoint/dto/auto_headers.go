package dto

type AutoHeaders struct {
	Disabled []string `json:"disabled"`
}

func NewAutoHeaders(disabled []string) AutoHeaders {
	return AutoHeaders{Disabled: append([]string{}, disabled...)}
}
