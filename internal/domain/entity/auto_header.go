package entity

const (
	HeaderAuthorization  = "Authorization"
	HeaderContentType    = "Content-Type"
	HeaderContentLength  = "Content-Length"
	HeaderHost           = "Host"
	HeaderUserAgent      = "User-Agent"
	HeaderAccept         = "Accept"
	HeaderAcceptEncoding = "Accept-Encoding"
	HeaderConnection     = "Connection"
	ComputedAtSend       = "<calculated when request is sent>"
)

type AutoHeader struct {
	Key      string
	Value    string
	Disabled bool
	Computed bool
}

func NewAutoHeader(key, value string) AutoHeader {
	return AutoHeader{Key: key, Value: value}
}

func NewComputedAutoHeader(key, value string) AutoHeader {
	return AutoHeader{Key: key, Value: value, Computed: true}
}

func (a AutoHeader) IsSent() bool {
	return !a.Disabled && !a.Computed
}

func (a AutoHeader) IsActive() bool {
	return !a.Disabled
}

func (a AutoHeader) LogValue() any {
	return map[string]any{"key": a.Key, "disabled": a.Disabled, "computed": a.Computed}
}
