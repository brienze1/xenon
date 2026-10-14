package entity

func (c ConsoleEntry) ActivityTarget() AIActivityTarget {
	return NewAIItemTarget(c.ID)
}
