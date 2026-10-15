package updatefailure

import "github.com/brienze1/xenon/internal/domain/enums"

const (
	None           enums.UpdateFailure = ""
	Unreachable    enums.UpdateFailure = "unreachable"
	DownloadFailed enums.UpdateFailure = "download_failed"
	Unverified     enums.UpdateFailure = "unverified"
	NotPackaged    enums.UpdateFailure = "not_packaged"
	NotWritable    enums.UpdateFailure = "not_writable"
	InstallFailed  enums.UpdateFailure = "install_failed"
)
