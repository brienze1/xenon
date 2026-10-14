package entity

func HasActiveWorkspace(localId string) bool {
	return localId != ""
}
