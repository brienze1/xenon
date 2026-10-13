package usecase

import (
	"sort"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/pmapi"
)

const reportedPmSegments = 3

func ScanPmUsage(scripts []entity.ScriptRef) []entity.PmUsage {
	occurrences := map[string][]entity.PmOccurrence{}
	for _, script := range scripts {
		for _, found := range pmChains(script.Text) {
			if isInMvpSubset(found.chain) {
				continue
			}
			api := reportedApi(found.chain)
			occurrences[api] = append(occurrences[api], entity.PmOccurrence{ItemID: script.ItemID, Path: script.Path, Phase: script.Phase, Line: found.line})
		}
	}
	apis := make([]string, 0, len(occurrences))
	for api := range occurrences {
		apis = append(apis, api)
	}
	sort.Strings(apis)
	usages := make([]entity.PmUsage, 0, len(apis))
	for _, api := range apis {
		usages = append(usages, entity.PmUsage{API: api, Occurrences: occurrences[api]})
	}
	return usages
}

func isInMvpSubset(chain string) bool {
	for _, allowed := range pmapi.MvpSubset {
		api := allowed.String()
		if chain == api || strings.HasPrefix(chain, api+".") || strings.HasPrefix(api, chain+".") {
			return true
		}
	}
	return false
}

func reportedApi(chain string) string {
	segments := strings.Split(chain, ".")
	if len(segments) > reportedPmSegments {
		segments = segments[:reportedPmSegments]
	}
	return strings.Join(segments, ".")
}

func isIdentifierStart(char byte) bool {
	return char == '_' || char == '$' || (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

func isIdentifierPart(char byte) bool {
	return isIdentifierStart(char) || (char >= '0' && char <= '9')
}
