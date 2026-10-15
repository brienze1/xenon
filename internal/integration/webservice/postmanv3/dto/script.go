package dto

const (
	preRequestListen   = "prerequest"
	postResponseListen = "test"
	scriptMimeType     = "text/javascript"
)

var scriptListens = map[string]string{
	"beforeRequest":      preRequestListen,
	"http:beforeRequest": preRequestListen,
	"beforeQuery":        preRequestListen,
	"afterResponse":      postResponseListen,
	"http:afterResponse": postResponseListen,
}

type Script struct {
	Extra    map[string]any `yaml:",inline"`
	Type     string         `yaml:"type"`
	Code     string         `yaml:"code"`
	Language string         `yaml:"language"`
}
