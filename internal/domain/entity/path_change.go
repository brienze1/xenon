package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type PathChange struct {
	Path   string
	Change enums.ChangeType
}
