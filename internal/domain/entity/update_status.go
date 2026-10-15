package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/updatefailure"
	"github.com/brienze1/xenon/internal/domain/enums/updatestate"
)

type UpdateStatus struct {
	State          enums.UpdateState
	Reason         enums.UpdateFailure
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
	Folder         string
	Received       int64
	Total          int64
	Automatic      bool
}

func (u UpdateStatus) HasState() bool {
	return u.State != updatestate.None
}

func (u UpdateStatus) IsBusy() bool {
	return u.State == updatestate.Checking || u.State == updatestate.Downloading || u.State == updatestate.Installing
}

func (u UpdateStatus) IsAvailable() bool {
	return u.State == updatestate.Available
}

func (u UpdateStatus) IsReady() bool {
	return u.State == updatestate.Ready
}

func (u UpdateStatus) CanDownload() bool {
	if u.State == updatestate.Available {
		return true
	}
	return u.State == updatestate.Failed && (u.Reason == updatefailure.DownloadFailed || u.Reason == updatefailure.Unverified || u.Reason == updatefailure.InstallFailed)
}

func (u UpdateStatus) HasReleasePage() bool {
	return u.ReleaseURL != ""
}

func (u UpdateStatus) with(state enums.UpdateState) UpdateStatus {
	u.State, u.Reason, u.Folder = state, updatefailure.None, ""
	return u
}

func (u UpdateStatus) LogValue() any {
	return map[string]any{
		"state":           u.State.String(),
		"reason":          u.Reason.String(),
		"current_version": u.CurrentVersion,
		"latest_version":  u.LatestVersion,
		"received":        u.Received,
		"total":           u.Total,
		"automatic":       u.Automatic,
	}
}
