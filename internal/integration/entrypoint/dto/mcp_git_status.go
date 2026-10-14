package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPGitStatus struct {
	Error      *SyncError `json:"error,omitempty"`
	LastSyncAt *string    `json:"last_sync_at,omitempty"`
	State      string     `json:"state"`
	Branch     string     `json:"branch,omitempty"`
	Remote     string     `json:"remote,omitempty"`
	Subfolder  string     `json:"subfolder,omitempty"`
	AccountId  string     `json:"account_id,omitempty"`
	Notice     string     `json:"notice,omitempty"`
	Ahead      int        `json:"ahead,omitempty"`
	Behind     int        `json:"behind,omitempty"`
	Changes    int        `json:"changes,omitempty"`
	Linked     bool       `json:"linked"`
	AutoPull   bool       `json:"auto_pull,omitempty"`
	AutoPush   bool       `json:"auto_push,omitempty"`
}

func NewMCPGitStatus(status entity.SyncStatus) MCPGitStatus {
	converted := NewSyncStatus(status)
	return MCPGitStatus{
		Error:      converted.LastError,
		LastSyncAt: converted.LastSyncAt,
		State:      converted.State,
		Branch:     status.Settings.Branch,
		Remote:     status.Settings.Remote(),
		Subfolder:  status.Settings.Subfolder,
		AccountId:  status.Settings.AccountID,
		Notice:     status.Notice,
		Ahead:      status.Ahead,
		Behind:     status.Behind,
		Changes:    status.LocalChanges,
		Linked:     status.IsLinked(),
		AutoPull:   status.Settings.AutoPull,
		AutoPush:   status.Settings.AutoPush,
	}
}
