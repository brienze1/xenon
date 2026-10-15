package entity

const releaseChecksumsName = "checksums.txt"

type Release struct {
	PageURL string
	Assets  []ReleaseAsset
	Version Version
}

func (r Release) PackageFor(platform UpdatePlatform) (ReleaseAsset, bool) {
	return r.asset("xenon-" + r.Version.String() + "-" + string(platform) + "." + platform.ArchiveExtension())
}

func (r Release) Checksums() (ReleaseAsset, bool) {
	return r.asset(releaseChecksumsName)
}

func (r Release) asset(name string) (ReleaseAsset, bool) {
	for _, asset := range r.Assets {
		if asset.IsNamed(name) {
			return asset, true
		}
	}
	return ReleaseAsset{}, false
}

func (r Release) LogValue() any {
	return map[string]any{"version": r.Version.String(), "assets": len(r.Assets)}
}
