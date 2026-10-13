package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type Folder struct {
	Seq     *int
	Name    string
	Docs    string
	File    string
	Headers []Pair
	Auth    Auth
	Scripts Scripts
}

func NewFolder(name string) Folder {
	return Folder{Name: name}
}

func (f *Folder) ToEntity(conversion *Conversion, path string) entity.Collection {
	folder := entity.NewFolder(conversion.newId(), f.Name)
	folder.Description = f.Docs
	folder.Auth = f.Auth.ToEntity()
	folder.Headers = keyValues(f.Headers)
	if f.Seq != nil {
		folder.Bruno = entity.SourceFields{"seq": *f.Seq}
	}
	conversion.checkAuth(f.Auth, folder.ID, path)
	conversion.addScripts(f.Scripts, folder.ID, path, f.File)
	return folder
}
