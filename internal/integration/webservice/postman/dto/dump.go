package dto

import (
	"sort"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	dumpManifest        = "archive.json"
	dumpCollectionsPart = "collection"
	dumpEnvironmentPart = "environment"
)

type Dump struct {
	Collections  []entity.ImportDocument
	Environments []entity.ImportDocument
	Entries      []string
}

func NewDump(documents []entity.ImportDocument, entries []string) Dump {
	dump := Dump{Collections: []entity.ImportDocument{}, Environments: []entity.ImportDocument{}, Entries: entries}
	for _, document := range documents {
		switch dumpPart(document.Path) {
		case dumpCollectionsPart:
			dump.Collections = append(dump.Collections, document)
		case dumpEnvironmentPart:
			dump.Environments = append(dump.Environments, document)
		}
	}
	return dump
}

func (d *Dump) ToEntity() entity.PostmanDump {
	return entity.PostmanDump{Collections: d.Collections, Environments: d.Environments, Skipped: d.skippedParts()}
}

func (d *Dump) skippedParts() []entity.PostmanDumpPart {
	files := map[string]int{}
	for _, entry := range d.Entries {
		part := dumpPart(entry)
		if entry != dumpManifest && part != dumpCollectionsPart && part != dumpEnvironmentPart {
			files[part]++
		}
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]entity.PostmanDumpPart, 0, len(names))
	for _, name := range names {
		parts = append(parts, entity.PostmanDumpPart{Name: name, Files: files[name]})
	}
	return parts
}

func dumpPart(path string) string {
	part, _, _ := strings.Cut(path, "/")
	return part
}
