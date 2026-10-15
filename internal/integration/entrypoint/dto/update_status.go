package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type UpdateStatus struct {
	State          string `json:"state"`
	Reason         string `json:"reason"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	ReleaseURL     string `json:"release_url"`
	Folder         string `json:"folder"`
	Received       int64  `json:"received"`
	Total          int64  `json:"total"`
	Automatic      bool   `json:"automatic"`
}

func NewUpdateStatus(status entity.UpdateStatus) UpdateStatus {
	return UpdateStatus{
		State:          status.State.String(),
		Reason:         status.Reason.String(),
		CurrentVersion: status.CurrentVersion,
		LatestVersion:  status.LatestVersion,
		ReleaseURL:     status.ReleaseURL,
		Folder:         status.Folder,
		Received:       status.Received,
		Total:          status.Total,
		Automatic:      status.Automatic,
	}
}

func (u UpdateStatus) LogValue() any {
	return map[string]any{"state": u.State, "reason": u.Reason, "current_version": u.CurrentVersion, "latest_version": u.LatestVersion}
}
