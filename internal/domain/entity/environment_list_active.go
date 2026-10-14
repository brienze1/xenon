package entity

func (e EnvironmentList) IsActive(id string) bool {
	return e.ActiveID == id
}
