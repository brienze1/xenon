package adapter

type Trash interface {
	MoveToTrash(path string) error
}
