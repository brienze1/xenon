package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const RequestChangedOnDiskEvent = "request:changed-on-disk"

type RequestChangedOnDisk struct {
	LocalId   string `json:"local_id"`
	RequestId string `json:"request_id"`
	Revision  string `json:"revision"`
	Deleted   bool   `json:"deleted"`
}

func NewRequestChangedOnDisk(localId string, change entity.RequestChange) RequestChangedOnDisk {
	return RequestChangedOnDisk{LocalId: localId, RequestId: change.RequestID, Revision: change.Revision, Deleted: change.Deleted}
}
