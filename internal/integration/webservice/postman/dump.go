package postman

import (
	"context"
	"sort"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	dumpManifest        = "archive.json"
	dumpCollectionsPart = "collection"
	dumpEnvironmentPart = "environment"
)

func (p *postmanWebservice) ReadDump(ctx context.Context, input entity.ImportInput) (*entity.PostmanDump, error) {
	logger.Debug(ctx, "Started", map[string]any{"root": input.Probe.Name, "documents": len(input.Documents)})

	dump := entity.PostmanDump{Collections: []entity.ImportDocument{}, Environments: []entity.ImportDocument{}, Skipped: []entity.PostmanDumpPart{}}
	for _, document := range input.Documents {
		switch dumpPart(document.Path) {
		case dumpCollectionsPart:
			dump.Collections = append(dump.Collections, document)
		case dumpEnvironmentPart:
			dump.Environments = append(dump.Environments, document)
		}
	}
	dump.Skipped = skippedParts(input.Probe.Entries)

	logger.Debug(ctx, "Finished", map[string]any{"root": input.Probe.Name, "collections": len(dump.Collections), "environments": len(dump.Environments), "skipped": len(dump.Skipped)})
	return &dump, nil
}

func skippedParts(entries []string) []entity.PostmanDumpPart {
	files := map[string]int{}
	for _, entry := range entries {
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
