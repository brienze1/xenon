package dto

type ShowInFolder struct {
	Path string `json:"path" validate:"required,absolute_path"`
}
