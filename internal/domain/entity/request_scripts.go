package entity

func (r Request) UsesOwnScripts() bool {
	return r.ScriptsEdited || !IsSavedRequestID(r.ID)
}
