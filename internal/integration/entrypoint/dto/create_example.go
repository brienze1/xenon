package dto

type CreateExample struct {
	LocalId   string `json:"local_id" validate:"required,not_blank"`
	RequestId string `json:"request_id" validate:"required,not_blank"`
	Name      string `json:"name"`
}
