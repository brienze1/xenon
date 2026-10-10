package entity

import (
	"path"
	"slices"
	"strings"
)

const (
	curlPrefix         = "curl "
	curlExePrefix      = "curl.exe "
	brunoFileExtension = ".bru"
	postmanV21Marker   = "/collection/v2.1"
	postmanV20Marker   = "/collection/v2.0"
	postmanV1Marker    = "/collection/v1"
	postmanV3Suffix    = ".request.yaml"
	dumpMarker         = "archive.json"
	brunoMarker        = "bruno.json"
	environmentScope   = "environment"
	globalsScope       = "globals"
	postmanV21         = "2.1"
	postmanV20         = "2.0"
)

var (
	openCollectionMarkers = []string{"opencollection.yml", "opencollection.yaml"}
	yamlFileExtensions    = []string{".yml", ".yaml"}
)

type ImportProbe struct {
	Name        string
	Text        string
	Schema      string
	Scope       string
	Entries     []string
	IsFolder    bool
	IsPasted    bool
	IsJson      bool
	IsMalformed bool
	HasItems    bool
	HasV1Shape  bool
}

func (i ImportProbe) Pasted(text string) ImportProbe {
	i.IsPasted = true
	i.Text = text
	return i
}

func (i ImportProbe) IsPastedCurl() bool {
	command := strings.ToLower(strings.TrimSpace(i.Text))
	return i.IsPasted && (strings.HasPrefix(command, curlPrefix) || strings.HasPrefix(command, curlExePrefix))
}

func (i ImportProbe) IsPostmanV3File() bool {
	return isPostmanV3Name(i.Name)
}

func (i ImportProbe) HasPostmanV3Entry() bool {
	return slices.ContainsFunc(i.Entries, isPostmanV3Name)
}

func (i ImportProbe) IsPostmanDump() bool {
	return slices.Contains(i.topLevelEntries(), dumpMarker)
}

func (i ImportProbe) IsBrunoBru() bool {
	return slices.Contains(i.topLevelEntries(), brunoMarker)
}

func (i ImportProbe) IsBrunoYaml() bool {
	topLevel := i.topLevelEntries()
	return slices.ContainsFunc(openCollectionMarkers, func(marker string) bool { return slices.Contains(topLevel, marker) })
}

func (i ImportProbe) MixesBrunoFormats() bool {
	mixesBruno := i.IsBrunoBru() && (i.IsBrunoYaml() || i.hasEntryWithExtension(yamlFileExtensions...))
	mixesYaml := i.IsBrunoYaml() && i.hasEntryWithExtension(brunoFileExtension)
	return mixesBruno || mixesYaml
}

func (i ImportProbe) PostmanV2Version() (string, bool) {
	switch {
	case strings.Contains(i.Schema, postmanV21Marker):
		return postmanV21, true
	case strings.Contains(i.Schema, postmanV20Marker):
		return postmanV20, true
	default:
		return "", false
	}
}

func (i ImportProbe) IsPostmanV1() bool {
	return i.HasV1Shape || strings.Contains(i.Schema, postmanV1Marker)
}

func (i ImportProbe) IsEnvironment() bool {
	return i.Scope == environmentScope
}

func (i ImportProbe) IsGlobals() bool {
	return i.Scope == globalsScope
}

func (i ImportProbe) topLevelEntries() []string {
	topLevel := make([]string, 0, len(i.Entries))
	for _, entry := range i.Entries {
		if !strings.Contains(entry, "/") {
			topLevel = append(topLevel, strings.ToLower(path.Base(entry)))
		}
	}
	return topLevel
}

func (i ImportProbe) hasEntryWithExtension(extensions ...string) bool {
	for _, entry := range i.Entries {
		if !isHiddenEntry(entry) && slices.Contains(extensions, strings.ToLower(path.Ext(entry))) {
			return true
		}
	}
	return false
}

func isPostmanV3Name(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), postmanV3Suffix)
}

func isHiddenEntry(entry string) bool {
	return slices.ContainsFunc(strings.Split(entry, "/"), func(segment string) bool { return strings.HasPrefix(segment, ".") })
}
