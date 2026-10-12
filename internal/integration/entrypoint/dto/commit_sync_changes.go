package dto

type CommitSyncChanges struct {
	LocalId string   `json:"local_id" validate:"required,not_blank"`
	Message string   `json:"message" validate:"required,not_blank"`
	ItemIDs []string `json:"item_ids" validate:"required,min=1,dive,required"`
}
