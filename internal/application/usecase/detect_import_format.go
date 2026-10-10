package usecase

import (
	"net/http"
	"path"
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/importkind"
	"github.com/brienze1/xenon/pkg/errs"
)

const (
	curlPrefix           = "curl "
	postmanV21Marker     = "/collection/v2.1"
	postmanV20Marker     = "/collection/v2.0"
	postmanV1Marker      = "/collection/v1"
	postmanV3Suffix      = ".request.yaml"
	dumpMarker           = "archive.json"
	brunoMarker          = "bruno.json"
	environmentScope     = "environment"
	globalsScope         = "globals"
	postmanV21           = "2.1"
	postmanV20           = "2.0"
	unsupportedFormatMsg = " is not a Postman, Bruno or cURL export"
)

var openCollectionMarkers = []string{"opencollection.yml", "opencollection.yaml"}

func DetectImportFormat(probe entity.ImportProbe) (enums.ImportKind, string, error) {
	if probe.IsPasted && strings.HasPrefix(strings.ToLower(strings.TrimSpace(probe.Text)), curlPrefix) {
		return importkind.Curl, "", nil
	}
	if probe.IsFolder {
		return detectFolder(probe)
	}
	if strings.HasSuffix(strings.ToLower(probe.Name), postmanV3Suffix) {
		return "", "", postmanV3Error()
	}
	if probe.IsMalformed {
		return "", "", errs.BadRequestError(probe.Name+" is not valid JSON", "IMP-01400")
	}
	if probe.IsJson {
		return detectJson(probe)
	}
	return "", "", unsupportedFormatError(probe.Name)
}

func detectFolder(probe entity.ImportProbe) (enums.ImportKind, string, error) {
	topLevel := topLevelEntries(probe.Entries)
	hasBruno := slices.Contains(topLevel, brunoMarker)
	hasOpenCollection := slices.ContainsFunc(openCollectionMarkers, func(marker string) bool { return slices.Contains(topLevel, marker) })
	switch {
	case slices.Contains(topLevel, dumpMarker):
		return importkind.PostmanDump, "", nil
	case hasBruno && hasOpenCollection:
		return "", "", errs.ClientError(probe.Name+" mixes Bruno .bru and YAML files", "IMP-00415", http.StatusUnsupportedMediaType)
	case hasBruno:
		return importkind.BrunoBru, "", nil
	case hasOpenCollection:
		return importkind.BrunoYaml, "", nil
	case slices.ContainsFunc(probe.Entries, func(entry string) bool { return strings.HasSuffix(strings.ToLower(entry), postmanV3Suffix) }):
		return "", "", postmanV3Error()
	default:
		return "", "", unsupportedFormatError(probe.Name)
	}
}

func detectJson(probe entity.ImportProbe) (enums.ImportKind, string, error) {
	switch {
	case strings.Contains(probe.Schema, postmanV21Marker):
		return importkind.PostmanCollection, postmanV21, nil
	case strings.Contains(probe.Schema, postmanV20Marker):
		return importkind.PostmanCollection, postmanV20, nil
	case probe.HasV1Shape || strings.Contains(probe.Schema, postmanV1Marker):
		return "", "", errs.BadRequestError("Postman v1 collections are not supported; export the collection as v2.1", "IMP-01400")
	case probe.Scope == environmentScope:
		return importkind.PostmanEnvironment, "", nil
	case probe.Scope == globalsScope:
		return importkind.PostmanGlobals, "", nil
	default:
		return "", "", unsupportedFormatError(probe.Name)
	}
}

func topLevelEntries(entries []string) []string {
	topLevel := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !strings.Contains(entry, "/") {
			topLevel = append(topLevel, strings.ToLower(path.Base(entry)))
		}
	}
	return topLevel
}

func postmanV3Error() error {
	return errs.ClientError("Postman v3 YAML is not supported yet; export the collection as v2.1 JSON", "IMP-00415", http.StatusUnsupportedMediaType)
}

func unsupportedFormatError(name string) error {
	return errs.ClientError(name+unsupportedFormatMsg, "IMP-00415", http.StatusUnsupportedMediaType)
}
