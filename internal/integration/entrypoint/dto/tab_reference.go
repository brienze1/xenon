package dto

type TabReference struct {
	TabId string `json:"tab_id" validate:"required,not_blank"`
}
