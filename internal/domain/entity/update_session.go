package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums/updatefailure"
	"github.com/brienze1/xenon/internal/domain/enums/updatestate"
)

type UpdateSession struct {
	Release        *Release
	PackagePath    string
	Status         UpdateStatus
	StartupChecked bool
}

func (u UpdateSession) For(installation Installation) UpdateSession {
	u.Status.CurrentVersion = installation.Version
	if !u.Status.HasState() {
		u.Status.State = updatestate.Idle
	}
	return u
}

func (u UpdateSession) Checking() (UpdateSession, bool) {
	if u.Status.IsBusy() {
		return u, false
	}
	u.Status = u.Status.with(updatestate.Checking)
	return u, true
}

func (u UpdateSession) StartingUp() (UpdateSession, bool) {
	if u.StartupChecked || u.Status.IsBusy() {
		return u, false
	}
	u.StartupChecked = true
	u.Status = u.Status.with(updatestate.Checking)
	return u, true
}

func (u UpdateSession) Checked(release Release, installation Installation) UpdateSession {
	u.Release, u.PackagePath = &release, ""
	u.Status.LatestVersion, u.Status.ReleaseURL = release.Version.String(), release.PageURL
	u.Status.Received, u.Status.Total, u.Status.Automatic = 0, 0, false
	_, hasPackage := release.PackageFor(installation.Platform)
	switch {
	case !release.Version.NewerThan(installation.RunningVersion()):
		u.Status = u.Status.with(updatestate.UpToDate)
	case !hasPackage:
		u.Status = u.Status.with(updatestate.NoPackage)
	default:
		u.Status = u.Status.with(updatestate.Available)
	}
	return u
}

func (u UpdateSession) CheckFailed() UpdateSession {
	u.Status = u.Status.with(updatestate.CheckFailed)
	u.Status.Reason = updatefailure.Unreachable
	return u
}

func (u UpdateSession) Downloading(automatic bool) (UpdateSession, bool) {
	if !u.Status.CanDownload() || u.Release == nil {
		return u, false
	}
	u.Status = u.Status.with(updatestate.Downloading)
	u.Status.Automatic, u.Status.Received, u.Status.Total, u.PackagePath = automatic, 0, 0, ""
	return u, true
}

func (u UpdateSession) Progressed(received, total int64) UpdateSession {
	u.Status.Received, u.Status.Total = received, total
	return u
}

func (u UpdateSession) Downloaded(path string) UpdateSession {
	u.PackagePath = path
	u.Status = u.Status.with(updatestate.Ready)
	return u
}

func (u UpdateSession) Installing() (UpdateSession, bool) {
	if !u.Status.IsReady() || u.PackagePath == "" {
		return u, false
	}
	u.Status = u.Status.with(updatestate.Installing)
	return u, true
}

func (u UpdateSession) Failed(failure *UpdateFailure) UpdateSession {
	u.Status = u.Status.with(updatestate.Failed)
	u.Status.Reason, u.Status.Folder = failure.Reason, failure.Folder
	return u
}

func (u UpdateSession) PackageFor(installation Installation) (ReleaseAsset, bool) {
	if u.Release == nil {
		return ReleaseAsset{}, false
	}
	return u.Release.PackageFor(installation.Platform)
}

func (u UpdateSession) Checksums() (ReleaseAsset, bool) {
	if u.Release == nil {
		return ReleaseAsset{}, false
	}
	return u.Release.Checksums()
}

func (u UpdateSession) LogValue() any {
	return map[string]any{"status": u.Status.LogValue(), "package_path": u.PackagePath, "startup_checked": u.StartupChecked}
}
