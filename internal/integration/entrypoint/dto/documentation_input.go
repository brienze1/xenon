package dto

type DocumentationInput struct {
	LocalId          string `json:"local_id" validate:"required,not_blank"`
	Id               string `json:"id" validate:"required,not_blank"`
	Yaml             string `json:"yaml"`
	ConfirmDeletions bool   `json:"confirm_deletions"`
	KeepDeleted      bool   `json:"keep_deleted"`
}

func (d DocumentationInput) LogValue() any {
	return map[string]any{"local_id": d.LocalId, "id": d.Id, "bytes": len(d.Yaml), "confirm_deletions": d.ConfirmDeletions, "keep_deleted": d.KeepDeleted}
}
