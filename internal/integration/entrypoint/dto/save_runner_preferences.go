package dto

type SaveRunnerPreferences struct {
	LocalId string    `json:"local_id" validate:"required,not_blank"`
	ItemId  string    `json:"item_id" validate:"required,not_blank"`
	Config  RunConfig `json:"config"`
}
