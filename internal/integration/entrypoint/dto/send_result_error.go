package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SendResultError struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Hint    string `json:"hint"`
}

func NewSendResultError(sendError entity.SendError) SendResultError {
	return SendResultError{Kind: sendError.Kind.String(), Message: sendError.Message, Hint: sendError.Hint}
}
