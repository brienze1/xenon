package entity

type ReleaseAsset struct {
	Name string
	URL  string
	Size int64
}

func (r ReleaseAsset) IsNamed(name string) bool {
	return r.Name == name
}

func (r ReleaseAsset) ExpectedSize(reported int64) int64 {
	if reported > 0 {
		return reported
	}
	return r.Size
}

func (r ReleaseAsset) LogValue() any {
	return map[string]any{"name": r.Name, "size": r.Size}
}
