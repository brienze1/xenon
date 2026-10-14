package entity

type AIApprovalEnvironment struct {
	ID    string
	Name  string
	Color string
}

func NewAIApprovalEnvironment(id, name, color string) AIApprovalEnvironment {
	return AIApprovalEnvironment{ID: id, Name: name, Color: color}
}
