package service

import (
	"context"
	"errors"
	"testing"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

type fakeUpdateSteps struct {
	session     entity.UpdateSession
	release     *entity.Release
	settingsErr error
	openErr     error
}

func (f *fakeUpdateSteps) FindUpdateSession(context.Context) entity.UpdateSession {
	return f.session
}

func (f *fakeUpdateSteps) SaveUpdateSession(_ context.Context, session entity.UpdateSession) {
	f.session = session
}

func (f *fakeUpdateSteps) ChangeUpdateSession(_ context.Context, change func(entity.UpdateSession) (entity.UpdateSession, bool)) (entity.UpdateSession, bool) {
	changed, accepted := change(f.session)
	if accepted {
		f.session = changed
	}
	return f.session, accepted
}

func (f *fakeUpdateSteps) PublishUpdateStatus(context.Context, entity.UpdateStatus) {}

func (f *fakeUpdateSteps) FindLatestRelease(context.Context) (*entity.Release, error) {
	return f.release, nil
}

func (f *fakeUpdateSteps) FindReleaseChecksums(context.Context, entity.ReleaseAsset) (entity.ReleaseChecksums, error) {
	return nil, nil
}

func (f *fakeUpdateSteps) DownloadUpdatePackage(context.Context, entity.ReleaseAsset, entity.ReleaseChecksums, func(received, total int64)) (string, error) {
	return "", nil
}

func (f *fakeUpdateSteps) InstallUpdate(context.Context, entity.Installation, string) error {
	return nil
}

func (f *fakeUpdateSteps) CleanUpdateLeftovers(context.Context, entity.Installation) {}

func (f *fakeUpdateSteps) RestartApp(context.Context, string) error {
	return nil
}

func (f *fakeUpdateSteps) OpenReleasePage(context.Context, string) error {
	return f.openErr
}

func (f *fakeUpdateSteps) FindAppSettings(context.Context) (*entity.AppSettings, error) {
	return nil, f.settingsErr
}

func newManageUpdateWith(steps *fakeUpdateSteps) adapter.ManageUpdateService {
	return NewManageUpdateService(steps, steps, steps, steps, steps, steps, steps, steps, steps, steps, steps)
}

func updateInstallation() entity.Installation {
	return entity.Installation{Version: "0.1.0", Path: "/apps/xenon", Platform: "linux-amd64", Installable: true}
}

func TestOpenReleasePagePassesTheOpenErrorThrough(t *testing.T) {
	failure := errs.ServiceUnavailableError(errors.New("no browser"), "TST-00503", "failed")
	steps := &fakeUpdateSteps{openErr: failure}
	steps.session.Status.ReleaseURL = "https://github.com/brienze1/xenon/releases/tag/v0.2.0"

	status, err := newManageUpdateWith(steps).OpenReleasePage(context.Background(), updateInstallation())

	if err != failure || status.ReleaseURL != steps.session.Status.ReleaseURL {
		t.Fatalf("OpenReleasePage() = %+v, %v; want the status and %v", status, err, failure)
	}
}

func TestCheckOnStartupPassesTheSettingsErrorThrough(t *testing.T) {
	failure := errs.InternalServerError(errors.New("unreadable"), "TST-01500", "failed")
	release := &entity.Release{
		Version: entity.Version{Minor: 2},
		Assets:  []entity.ReleaseAsset{{Name: "xenon-0.2.0-linux-amd64.tar.gz"}},
	}
	steps := &fakeUpdateSteps{release: release, settingsErr: failure}

	status, err := newManageUpdateWith(steps).CheckOnStartup(context.Background(), updateInstallation())

	if err != failure || !status.IsAvailable() {
		t.Fatalf("CheckOnStartup() = %+v, %v; want an available update and %v", status, err, failure)
	}
}
