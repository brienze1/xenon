package entity

func ChangeItemIDs(changes []ChangeItem) []string {
	ids := make([]string, 0, len(changes))
	for _, change := range changes {
		ids = append(ids, change.ItemID)
	}
	return ids
}
