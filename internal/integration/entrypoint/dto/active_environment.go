package dto

type ActiveEnvironment struct {
	ActiveId string `json:"active_id"`
}

func NewActiveEnvironment(activeId string) ActiveEnvironment {
	return ActiveEnvironment{ActiveId: activeId}
}
