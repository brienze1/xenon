package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RevealedSecret struct {
	Value string `json:"value"`
}

func NewRevealedSecret(value string) RevealedSecret {
	return RevealedSecret{Value: value}
}

func (r RevealedSecret) LogValue() any {
	return map[string]any{"value": entity.MaskedValue}
}
