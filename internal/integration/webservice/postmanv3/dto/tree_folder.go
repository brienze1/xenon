package dto

type TreeFolder struct {
	Definition *TreeDocument
	Subfolders []string
	Requests   []TreeDocument
}
