package dto

import "github.com/brienze1/xenon/internal/domain/entity"

func NewRunRowSendResult(runId string, index int, response entity.Response) SendResult {
	return NewLastSendResult(entity.RunResponseKey(runId, index), response)
}
