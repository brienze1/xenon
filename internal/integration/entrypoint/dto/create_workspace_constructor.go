package dto

func NewCreateWorkspace(name, path string) CreateWorkspace {
	return CreateWorkspace{Name: name, Path: path}
}
