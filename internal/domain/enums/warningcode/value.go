package warningcode

import "github.com/brienze1/xenon/internal/domain/enums"

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
	LiteralCredential      enums.WarningCode = "literal_credential"
	ProtocolScript         enums.WarningCode = "protocol_script"
	LinkedSecret           enums.WarningCode = "linked_secret"
	Unknown                enums.WarningCode = "unknown"
)
