package entity

func (e Environment) Target() AIActivityTarget {
	return AIActivityTarget{ID: e.ID, Name: e.Name}
}

func NewAIEnvironmentTarget(id string) AIActivityTarget {
	return AIActivityTarget{ID: id}
}
