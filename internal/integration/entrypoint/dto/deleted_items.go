package dto

type DeletedItems struct {
	DeletedRequestIds []string `json:"deleted_request_ids"`
}

func NewDeletedItems(requestIds []string) DeletedItems {
	return DeletedItems{DeletedRequestIds: append([]string{}, requestIds...)}
}
