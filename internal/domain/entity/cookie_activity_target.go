package entity

func (c Cookie) ActivityTarget() AIActivityTarget {
	return AIActivityTarget{Name: c.Name, Path: c.Domain + c.Path}
}
