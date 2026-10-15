package entity

import (
	"path/filepath"
	"strings"
)

const (
	bundleExecutableFolder = "Contents/MacOS"
	bundleExecutableName   = "xenon"
	bundleSuffix           = ".app"
	updateOldSuffix        = ".xenon-old"
	updateExtractSuffix    = ".xenon-update"
	updateProbeSuffix      = ".xenon-update-probe"
)

var plainExecutables = map[string]bool{"xenon": true, "xenon.exe": true}

type Installation struct {
	Version     string
	Path        string
	Platform    UpdatePlatform
	Installable bool
}

func (i Installation) RunningVersion() Version {
	version, err := ParseVersion(i.Version)
	if err != nil {
		return Version{}
	}
	return version
}

func (i Installation) CanBeReplaced() bool {
	return i.Installable && i.Path != ""
}

func (i Installation) HasPath() bool {
	return i.Path != ""
}

func (i Installation) Folder() string {
	return filepath.Dir(i.Path)
}

func (i Installation) Name() string {
	return filepath.Base(i.Path)
}

func (i Installation) Executable() string {
	if i.Platform.IsBundle() {
		return filepath.Join(i.Path, filepath.FromSlash(bundleExecutableFolder), bundleExecutableName)
	}
	return i.Path
}

func (i Installation) AcceptsPackageEntry(name string, isDir, isRegular bool) bool {
	if i.Platform.IsBundle() {
		return isDir && strings.HasSuffix(name, bundleSuffix)
	}
	return isRegular && plainExecutables[name]
}

func (i Installation) PreviousPath() string {
	return i.sidePath(updateOldSuffix)
}

func (i Installation) ExtractionPath() string {
	return i.sidePath(updateExtractSuffix)
}

func (i Installation) ProbePath() string {
	return i.sidePath(updateProbeSuffix)
}

func (i Installation) sidePath(suffix string) string {
	return filepath.Join(i.Folder(), "."+i.Name()+suffix)
}

func (i Installation) LogValue() any {
	return map[string]any{"version": i.Version, "path": i.Path, "platform": string(i.Platform), "installable": i.Installable}
}
