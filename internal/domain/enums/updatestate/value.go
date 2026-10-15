package updatestate

import "github.com/brienze1/xenon/internal/domain/enums"

const (
	Idle        enums.UpdateState = "idle"
	Checking    enums.UpdateState = "checking"
	UpToDate    enums.UpdateState = "up_to_date"
	Available   enums.UpdateState = "available"
	NoPackage   enums.UpdateState = "no_package"
	CheckFailed enums.UpdateState = "check_failed"
	Downloading enums.UpdateState = "downloading"
	Ready       enums.UpdateState = "ready"
	Installing  enums.UpdateState = "installing"
	Failed      enums.UpdateState = "failed"
)
