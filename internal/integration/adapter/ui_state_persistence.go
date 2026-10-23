package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type UIStatePersistence interface {
	usecase.FindUIState
	usecase.SaveUIState
	usecase.DeleteUIState
	usecase.FindDigestState
	usecase.SaveDigestState
	usecase.FindCookieAllowlist
	usecase.SaveCookieAllowlist
	usecase.FindRunnerPreferences
	usecase.SaveRunnerPreferences
	usecase.DeleteRunnerPreferences
	usecase.FindLocalSyncSettings
	usecase.SaveLocalSyncSettings
	usecase.LocateClone
}
