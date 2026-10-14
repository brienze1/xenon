package dto

func NewSaveResponseBody(tabId, path string) SaveResponseBody {
	return SaveResponseBody{TabId: tabId, Path: path}
}
