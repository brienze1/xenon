package warningcode

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	UnsupportedAuth        enums.WarningCode = "unsupported_auth"
	PmRequire              enums.WarningCode = "pm_require"
	MissingFile            enums.WarningCode = "missing_file"
	FolderVariables        enums.WarningCode = "folder_variables"
	RequestVariables       enums.WarningCode = "request_variables"
	UntranslatedScript     enums.WarningCode = "untranslated_script"
	SkippedPart            enums.WarningCode = "skipped_part"
	UnsupportedRequestType enums.WarningCode = "unsupported_request_type"
	DisabledBody           enums.WarningCode = "disabled_body"
	RawHeaderString        enums.WarningCode = "raw_header_string"
	PmApiOutsideMvp        enums.WarningCode = "pm_api_outside_mvp"
	MergedIntoFolder       enums.WarningCode = "merged_into_folder"
	WrappedInFolder        enums.WarningCode = "wrapped_in_folder"
	VariableKept           enums.WarningCode = "variable_kept"
	Unknown                enums.WarningCode = "unknown"
)

func GetWarningCode(s string) enums.WarningCode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "unsupported_auth":
		return UnsupportedAuth
	case "pm_require":
		return PmRequire
	case "missing_file":
		return MissingFile
	case "folder_variables":
		return FolderVariables
	case "request_variables":
		return RequestVariables
	case "untranslated_script":
		return UntranslatedScript
	case "skipped_part":
		return SkippedPart
	case "unsupported_request_type":
		return UnsupportedRequestType
	case "disabled_body":
		return DisabledBody
	case "raw_header_string":
		return RawHeaderString
	case "pm_api_outside_mvp":
		return PmApiOutsideMvp
	case "merged_into_folder":
		return MergedIntoFolder
	case "wrapped_in_folder":
		return WrappedInFolder
	case "variable_kept":
		return VariableKept
	default:
		return Unknown
	}
}
