package entity

const (
	duplicateIdWarning = "another item has the same id, so this one uses a new id that is written when it is saved"
	missingIdWarning   = "the file has no id, so it gets one that is written when it is saved"
)

type ItemIdentities struct {
	previous     map[string]string
	seen         map[string]string
	reidentified map[string]string
}

func NewItemIdentities(previous map[string]string) *ItemIdentities {
	return &ItemIdentities{previous: previous, seen: map[string]string{}, reidentified: map[string]string{}}
}

func (i *ItemIdentities) Identify(storedId, path string, newId func() string) (string, string) {
	if storedId != "" && i.seen[storedId] == "" {
		i.seen[storedId] = path
		return storedId, ""
	}
	id, known := i.previous[path]
	if !known || i.seen[id] != "" {
		id = newId()
	}
	i.seen[id] = path
	i.reidentified[path] = id
	if storedId == "" {
		return id, missingIdWarning
	}
	return id, duplicateIdWarning
}

func (i *ItemIdentities) Reidentified() map[string]string {
	return i.reidentified
}
