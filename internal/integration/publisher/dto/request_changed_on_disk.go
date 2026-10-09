package dto

type RequestChangedOnDisk struct {
	LocalId   string `json:"local_id"`
	RequestId string `json:"request_id"`
	Revision  string `json:"revision"`
	Deleted   bool   `json:"deleted"`
}
