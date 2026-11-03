package entity

func (e EnvironmentList) IsActive(id string) bool {
	return e.ActiveID == id
}

// WithActive is a copy that shows id as the active environment; an id that matches no environment shows none.
func (e EnvironmentList) WithActive(id string) EnvironmentList {
	e.ActiveID = ""
	if e.Contains(id) {
		e.ActiveID = id
	}
	return e
}
