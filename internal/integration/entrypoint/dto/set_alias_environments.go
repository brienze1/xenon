package dto

type SetAliasEnvironments struct {
	LocalId        string   `json:"local_id" validate:"required,not_blank"`
	RequestId      string   `json:"request_id" validate:"required,not_blank"`
	AliasId        string   `json:"alias_id" validate:"required,not_blank"`
	BaseRevision   string   `json:"base_revision"`
	EnvironmentIds []string `json:"environment_ids"`
}

func NewSetAliasEnvironments(localId string, saved Request, aliasId string, environmentIds []string) SetAliasEnvironments {
	return SetAliasEnvironments{LocalId: localId, RequestId: saved.Id, AliasId: aliasId, BaseRevision: saved.Revision, EnvironmentIds: environmentIds}
}
