package adapter

type Archive interface {
	Extract(archivePath, targetDir string, limit int64) error
}
