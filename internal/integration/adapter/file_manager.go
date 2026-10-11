package adapter

type FileManager interface {
	ShowInFolder(path string) error
}
