package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type AuthUnsupported struct {
	Data   any
	Origin enums.ImportOrigin
	Type   string
}

func (a AuthUnsupported) IsEmpty() bool {
	return a.Type == ""
}

func (a AuthUnsupported) LogValue() any {
	return map[string]any{"origin": a.Origin, "type": a.Type}
}
