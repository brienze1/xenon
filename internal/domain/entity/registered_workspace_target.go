package entity

func (r RegisteredWorkspace) Target() AIActivityTarget {
	return AIActivityTarget{ID: r.LocalID, Name: r.DisplayNameCache, Path: r.Path}
}

func (r RegisteredWorkspace) NamedTarget() AIActivityTarget {
	return AIActivityTarget{ID: r.LocalID, Name: r.DisplayNameCache}
}
