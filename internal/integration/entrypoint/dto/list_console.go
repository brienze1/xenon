package dto

type ListConsole struct {
	AfterId string `json:"after_id"`
	Limit   int    `json:"limit" validate:"required,min=1,max=5000"`
}
