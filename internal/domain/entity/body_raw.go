package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type BodyRaw struct {
	Language enums.RawLanguage
	Content  string
}
