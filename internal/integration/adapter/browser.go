package adapter

type Browser interface {
	OpenURL(url string) error
}
