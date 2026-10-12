package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SyncError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	NextStep string `json:"next_step"`
}

func NewSyncError(failure entity.SyncError) SyncError {
	return SyncError{Code: failure.Code, Message: failure.Message, NextStep: failure.NextStep}
}
