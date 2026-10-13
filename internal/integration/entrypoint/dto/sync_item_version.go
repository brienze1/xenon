package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SyncItemVersion struct {
	Request     *Request                `json:"request,omitempty"`
	Environment *SyncEnvironmentVersion `json:"environment,omitempty"`
	Text        string                  `json:"text"`
}

func NewSyncItemVersion(version *entity.ConflictVersion) *SyncItemVersion {
	if version == nil {
		return nil
	}
	converted := SyncItemVersion{Text: version.Text}
	if version.Request != nil {
		request := NewRequest(*version.Request)
		converted.Request = &request
	}
	if version.Environment != nil {
		environment := NewSyncEnvironmentVersion(*version.Environment)
		converted.Environment = &environment
	}
	return &converted
}
