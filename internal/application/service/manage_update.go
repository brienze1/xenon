package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/updatefailure"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageUpdate struct {
	findUpdateSession     usecase.FindUpdateSession
	saveUpdateSession     usecase.SaveUpdateSession
	publishUpdateStatus   usecase.PublishUpdateStatus
	findLatestRelease     usecase.FindLatestRelease
	findReleaseChecksums  usecase.FindReleaseChecksums
	downloadUpdatePackage usecase.DownloadUpdatePackage
	installUpdate         usecase.InstallUpdate
	cleanUpdateLeftovers  usecase.CleanUpdateLeftovers
	restartApp            usecase.RestartApp
	openReleasePage       usecase.OpenReleasePage
	findAppSettings       usecase.FindAppSettings
}

func NewManageUpdateService(
	findUpdateSession usecase.FindUpdateSession,
	saveUpdateSession usecase.SaveUpdateSession,
	publishUpdateStatus usecase.PublishUpdateStatus,
	findLatestRelease usecase.FindLatestRelease,
	findReleaseChecksums usecase.FindReleaseChecksums,
	downloadUpdatePackage usecase.DownloadUpdatePackage,
	installUpdate usecase.InstallUpdate,
	cleanUpdateLeftovers usecase.CleanUpdateLeftovers,
	restartApp usecase.RestartApp,
	openReleasePage usecase.OpenReleasePage,
	findAppSettings usecase.FindAppSettings,
) adapter.ManageUpdateService {
	return &manageUpdate{
		findUpdateSession:     findUpdateSession,
		saveUpdateSession:     saveUpdateSession,
		publishUpdateStatus:   publishUpdateStatus,
		findLatestRelease:     findLatestRelease,
		findReleaseChecksums:  findReleaseChecksums,
		downloadUpdatePackage: downloadUpdatePackage,
		installUpdate:         installUpdate,
		cleanUpdateLeftovers:  cleanUpdateLeftovers,
		restartApp:            restartApp,
		openReleasePage:       openReleasePage,
		findAppSettings:       findAppSettings,
	}
}

func (m *manageUpdate) Status(ctx context.Context, installation entity.Installation) entity.UpdateStatus {
	logger.Info(ctx, "Started", installation)

	status := m.current(ctx, installation)

	logger.Info(ctx, "Finished", installation, status)
	return status
}

func (m *manageUpdate) Check(ctx context.Context, installation entity.Installation) entity.UpdateStatus {
	logger.Info(ctx, "Started", installation)

	session, started := m.change(ctx, installation, entity.UpdateSession.Checking)
	if started {
		session = m.check(ctx, installation, session)
	}

	logger.Info(ctx, "Finished", installation, session.Status)
	return session.Status
}

func (m *manageUpdate) CheckOnStartup(ctx context.Context, installation entity.Installation) (entity.UpdateStatus, error) {
	logger.Info(ctx, "Started", installation)

	session, started := m.change(ctx, installation, entity.UpdateSession.StartingUp)
	if !started {
		logger.Info(ctx, "Finished - already checked", installation, session.Status)
		return session.Status, nil
	}
	m.cleanUpdateLeftovers.CleanUpdateLeftovers(ctx, installation)
	session = m.check(ctx, installation, session)
	if !session.Status.IsAvailable() {
		logger.Info(ctx, "Finished", installation, session.Status)
		return session.Status, nil
	}
	settings, err := m.findAppSettings.FindAppSettings(ctx)
	if err != nil {
		return session.Status, err
	}
	if !settings.UpdatesAutomatically() {
		logger.Info(ctx, "Finished", installation, session.Status)
		return session.Status, nil
	}
	session, started = m.change(ctx, installation, func(current entity.UpdateSession) (entity.UpdateSession, bool) {
		return current.Downloading(true)
	})
	if started {
		go m.download(context.WithoutCancel(ctx), installation, session)
	}

	logger.Info(ctx, "Finished - downloading in the background", installation, session.Status)
	return session.Status, nil
}

func (m *manageUpdate) Download(ctx context.Context, installation entity.Installation) entity.UpdateStatus {
	logger.Info(ctx, "Started", installation)

	session, started := m.change(ctx, installation, func(current entity.UpdateSession) (entity.UpdateSession, bool) {
		return current.Downloading(false)
	})
	if started {
		session = m.download(ctx, installation, session)
	}

	logger.Info(ctx, "Finished", installation, session.Status)
	return session.Status
}

func (m *manageUpdate) Restart(ctx context.Context, installation entity.Installation) entity.UpdateStatus {
	logger.Info(ctx, "Started", installation)

	session, started := m.change(ctx, installation, entity.UpdateSession.Installing)
	if !started {
		logger.Info(ctx, "Finished - no update ready", installation, session.Status)
		return session.Status
	}
	session, installed := m.install(ctx, installation, session)
	if installed {
		if err := m.restartApp.RestartApp(ctx, installation.Executable()); err != nil {
			session = m.fail(ctx, session, entity.UpdateFailureOf(err, updatefailure.InstallFailed))
		}
	}

	logger.Info(ctx, "Finished", installation, session.Status)
	return session.Status
}

func (m *manageUpdate) InstallOnQuit(ctx context.Context, installation entity.Installation) {
	logger.Info(ctx, "Started", installation)

	session, started := m.change(ctx, installation, entity.UpdateSession.Installing)
	if started {
		session, _ = m.install(ctx, installation, session)
	}

	logger.Info(ctx, "Finished", installation, session.Status)
}

func (m *manageUpdate) OpenReleasePage(ctx context.Context, installation entity.Installation) (entity.UpdateStatus, error) {
	logger.Info(ctx, "Started", installation)

	status := m.current(ctx, installation)
	if status.HasReleasePage() {
		if err := m.openReleasePage.OpenReleasePage(ctx, status.ReleaseURL); err != nil {
			return status, err
		}
	}

	logger.Info(ctx, "Finished", installation, status)
	return status, nil
}

func (m *manageUpdate) current(ctx context.Context, installation entity.Installation) entity.UpdateStatus {
	return m.findUpdateSession.FindUpdateSession(ctx).For(installation).Status
}

func (m *manageUpdate) check(ctx context.Context, installation entity.Installation, session entity.UpdateSession) entity.UpdateSession {
	release, err := m.findLatestRelease.FindLatestRelease(ctx)
	if err != nil {
		logger.Warn(ctx, err, "Couldn't check for updates", installation)
		return m.store(ctx, session.CheckFailed())
	}
	session = m.store(ctx, session.Checked(*release, installation))
	logger.Info(ctx, "Latest release checked", release, session.Status)
	return session
}

func (m *manageUpdate) download(ctx context.Context, installation entity.Installation, session entity.UpdateSession) entity.UpdateSession {
	asset, hasPackage := session.PackageFor(installation)
	checksumsAsset, hasChecksums := session.Checksums()
	if !hasPackage || !hasChecksums {
		return m.fail(ctx, session, entity.NewUpdateFailure(updatefailure.Unverified, nil))
	}
	checksums, err := m.findReleaseChecksums.FindReleaseChecksums(ctx, checksumsAsset)
	if err != nil {
		return m.fail(ctx, session, entity.UpdateFailureOf(err, updatefailure.DownloadFailed))
	}
	if !checksums.Lists(asset.Name) {
		return m.fail(ctx, session, entity.NewUpdateFailure(updatefailure.Unverified, nil))
	}
	path, err := m.downloadUpdatePackage.DownloadUpdatePackage(ctx, asset, checksums, func(received, total int64) {
		session = m.store(ctx, session.Progressed(received, total))
	})
	if err != nil {
		return m.fail(ctx, session, entity.UpdateFailureOf(err, updatefailure.DownloadFailed))
	}
	logger.Info(ctx, "Update downloaded and verified", asset)
	return m.store(ctx, session.Downloaded(path))
}

func (m *manageUpdate) install(ctx context.Context, installation entity.Installation, session entity.UpdateSession) (entity.UpdateSession, bool) {
	if !installation.CanBeReplaced() {
		return m.fail(ctx, session, entity.NewUpdateFailure(updatefailure.NotPackaged, nil)), false
	}
	if err := m.installUpdate.InstallUpdate(ctx, installation, session.PackagePath); err != nil {
		return m.fail(ctx, session, entity.UpdateFailureOf(err, updatefailure.InstallFailed)), false
	}
	logger.Info(ctx, "Update installed", installation, session.Status)
	return session, true
}

func (m *manageUpdate) fail(ctx context.Context, session entity.UpdateSession, failure *entity.UpdateFailure) entity.UpdateSession {
	logger.Warn(ctx, failure, "Update failed", session.Status)
	return m.store(ctx, session.Failed(failure))
}

func (m *manageUpdate) change(ctx context.Context, installation entity.Installation, change func(entity.UpdateSession) (entity.UpdateSession, bool)) (entity.UpdateSession, bool) {
	session, changed := m.saveUpdateSession.ChangeUpdateSession(ctx, func(current entity.UpdateSession) (entity.UpdateSession, bool) {
		return change(current.For(installation))
	})
	if changed {
		m.publish(ctx, session.Status)
	}
	return session.For(installation), changed
}

func (m *manageUpdate) store(ctx context.Context, session entity.UpdateSession) entity.UpdateSession {
	m.saveUpdateSession.SaveUpdateSession(ctx, session)
	m.publish(ctx, session.Status)
	return session
}

func (m *manageUpdate) publish(ctx context.Context, status entity.UpdateStatus) {
	m.publishUpdateStatus.PublishUpdateStatus(ctx, status)
}
