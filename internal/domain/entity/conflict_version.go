package entity

type ConflictVersion struct {
	Request     *Request
	Environment *Environment
	Text        string
}
