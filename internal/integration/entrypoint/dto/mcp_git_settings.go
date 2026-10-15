package dto

import (
	"cmp"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type MCPGitSettings struct {
	AutoPull  *bool  `json:"auto_pull" description:"Pull automatically; omitted keeps the stored value."`
	AutoPush  *bool  `json:"auto_push" description:"Push automatically; omitted keeps the stored value."`
	AccountId string `json:"account_id" description:"Git account id, from the Sync settings."`
}

// SaveSyncSettings keeps every stored option the arguments leave out.
func (m *MCPGitSettings) SaveSyncSettings(localId string, current entity.LocalSyncSettings) SaveSyncSettings {
	autoPull, autoPush := current.AutoPull, current.AutoPush
	if m.AutoPull != nil {
		autoPull = *m.AutoPull
	}
	if m.AutoPush != nil {
		autoPush = *m.AutoPush
	}
	return NewSaveSyncSettings(localId, current.WithOptions(cmp.Or(m.AccountId, current.AccountID), autoPull, autoPush))
}
