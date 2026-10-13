package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/sshmode"
)

type AddGitSsh struct {
	Mode        string `json:"mode" validate:"required,oneof=agent key_file generate"`
	KeyPath     string `json:"key_path" validate:"required_if=Mode key_file"`
	Passphrase  string `json:"passphrase"`
	Host        string `json:"host" validate:"required,not_blank"`
	DisplayName string `json:"display_name"`
	CommitName  string `json:"commit_name"`
	CommitEmail string `json:"commit_email" validate:"omitempty,email"`
}

func (a *AddGitSsh) ToEntity() entity.GitAccount {
	mode := sshmode.GetSshMode(a.Mode)
	if a.Mode == "generate" {
		mode = sshmode.Generated
	}
	return entity.GitAccount{SshMode: mode, Host: a.Host, DisplayName: a.DisplayName, CommitName: a.CommitName, CommitEmail: a.CommitEmail}
}

func (a AddGitSsh) LogValue() any {
	return map[string]any{"mode": a.Mode, "keyPath": a.KeyPath, "host": a.Host, "hasPassphrase": a.Passphrase != ""}
}
