package dto

type CreateAlias struct {
	LocalId        string   `json:"local_id" validate:"required,not_blank"`
	RequestId      string   `json:"request_id" validate:"required,not_blank"`
	BaseRevision   string   `json:"base_revision"`
	Name           string   `json:"name"`
	EnvironmentIds []string `json:"environment_ids"`
}

func NewCreateAlias(localId, requestId, name string, environmentIds []string) CreateAlias {
	return CreateAlias{LocalId: localId, RequestId: requestId, Name: name, EnvironmentIds: environmentIds}
}
