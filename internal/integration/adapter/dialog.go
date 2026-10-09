package adapter

type Dialog interface {
	PickFolder(title string) (string, error)
	PickFile(title string) (string, error)
	PickSaveFile(title, suggestedName string) (string, error)
}
