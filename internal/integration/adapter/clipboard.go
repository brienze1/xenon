package adapter

type Clipboard interface {
	SetText(text string) error
}
