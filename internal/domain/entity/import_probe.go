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
	postmanV3Folder    = "postman"
	v3CollectionsDir   = "collections"
	v3EnvironmentsDir  = "environments"
	v3GlobalsDir       = "globals"
	v3ResourcesDir     = ".resources"
	v3DefinitionFile   = "definition.yaml"
	v3CollectionKind   = "collection"
	v3HttpRequestKind  = "http-request"
	dumpMarker         = "archive.json"
	brunoMarker        = "bruno.json"
	environmentScope   = "environment"
	globalsScope       = "globals"
	postmanV21         = "2.1"
	postmanV20         = "2.0"
	xenonBundleKind    = "bundle"
	xenonExportKind    = "export"
	LatestXenonFormat  = 1
	PostmanV3Version   = "3"
)

var (
	openCollectionMarkers  = []string{"opencollection.yml", "opencollection.yaml"}
	yamlFileExtensions     = []string{".yml", ".yaml"}
	v3EnvironmentSuffixes  = []string{".environment.yaml", ".environment.yml", ".environment.json"}
	v3GlobalsSuffixes      = []string{".globals.yaml", ".globals.yml", ".globals.json"}
	v3WorkspaceDirectories = []string{v3CollectionsDir, v3EnvironmentsDir, v3GlobalsDir}
)

type ImportProbe struct {
	XenonKind         string
	XenonRoot         string
	Name              string
	Text              string
	Schema            string
	Scope             string
	PostmanKind       string
	DocumentID        string
	Entries           []string
	XenonFormat       int
	IsFolder          bool
	IsPasted          bool
	IsJson            bool
	IsYaml            bool
	IsMalformed       bool
	HasItems          bool
	HasV1Shape        bool
	HasValues         bool
	HasWorkspaceField bool
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

func (i ImportProbe) IsPostmanV3Request() bool {
	return !i.IsFolder && (isPostmanV3Name(i.Name) || i.PostmanKind == v3HttpRequestKind)
}

func (i ImportProbe) IsPostmanV3Collection() bool {
	return !i.IsFolder && i.Schema == "" && i.PostmanKind == v3CollectionKind
}

func (i ImportProbe) IsPostmanV3Variables() bool {
	return !i.IsFolder && i.HasValues && i.Scope == "" && i.Schema == "" && i.PostmanKind == ""
}

func (i ImportProbe) IsPostmanV3Globals() bool {
	return hasSuffixFold(i.Name, v3GlobalsSuffixes...) || i.HasWorkspaceField || hasSuffixFold(i.DocumentID, v3GlobalsSuffixes...)
}

func (i ImportProbe) IsPostmanV3CollectionFolder() bool {
	return i.IsFolder && slices.ContainsFunc(i.Entries, isPostmanV3Entry)
}

func (i ImportProbe) IsPostmanV3Workspace() bool {
	if !i.IsFolder {
		return false
	}
	if strings.EqualFold(i.Name, v3CollectionsDir) && !slices.ContainsFunc(i.topLevelV3Entries(), isPostmanV3Entry) {
		return slices.ContainsFunc(i.Entries, isPostmanV3Entry)
	}
	return slices.ContainsFunc(i.Entries, isPostmanV3WorkspaceEntry)
}

func (i ImportProbe) topLevelV3Entries() []string {
	return slices.DeleteFunc(slices.Clone(i.Entries), func(entry string) bool {
		segments := strings.Split(entry, "/")
		return len(segments) > 1 && !(len(segments) == 2 && segments[0] == v3ResourcesDir)
	})
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

func isPostmanV3Entry(entry string) bool {
	return isPostmanV3Name(entry) || strings.HasSuffix(entry, v3ResourcesDir+"/"+v3DefinitionFile)
}

func isPostmanV3WorkspaceEntry(entry string) bool {
	segments := strings.Split(entry, "/")
	if len(segments) > 1 && segments[0] == postmanV3Folder && slices.Contains(v3WorkspaceDirectories, segments[1]) {
		segments = segments[1:]
	}
	switch {
	case len(segments) >= 3 && segments[0] == v3CollectionsDir:
		return isPostmanV3Entry(entry)
	case len(segments) == 2 && segments[0] == v3EnvironmentsDir:
		return hasSuffixFold(segments[1], v3EnvironmentSuffixes...)
	case len(segments) == 2 && segments[0] == v3GlobalsDir:
		return hasSuffixFold(segments[1], v3GlobalsSuffixes...)
	default:
		return false
	}
}

func hasSuffixFold(name string, suffixes ...string) bool {
	lower := strings.ToLower(name)
	return slices.ContainsFunc(suffixes, func(suffix string) bool { return strings.HasSuffix(lower, suffix) })
}

func isHiddenEntry(entry string) bool {
	return slices.ContainsFunc(strings.Split(entry, "/"), func(segment string) bool { return strings.HasPrefix(segment, ".") })
}

func (i ImportProbe) IsXenon() bool {
	return i.XenonFormat > 0 && i.XenonKind != ""
}

func (i ImportProbe) IsNewerXenon() bool {
	return i.XenonFormat > LatestXenonFormat
}

func (i ImportProbe) IsXenonBundle() bool {
	return i.XenonKind == xenonBundleKind
}

func (i ImportProbe) IsXenonExport() bool {
	return i.XenonKind == xenonExportKind
}
