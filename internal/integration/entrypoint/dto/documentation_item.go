package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type DocumentationItem struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Folder   string `json:"folder"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	Revision string `json:"revision"`
}

func NewDocumentationItem(request entity.Request) DocumentationItem {
	return DocumentationItem{Id: request.ID, Name: request.Name, Method: request.MethodOrDefault(), Revision: request.Revision}
}

func NewCreatedDocumentationItem(created entity.CreatedRequest) DocumentationItem {
	item := NewDocumentationItem(created.Request)
	item.Folder, item.Path = created.Folder(), created.Path
	return item
}
