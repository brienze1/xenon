package entity

type ReleaseAsset struct {
	Name string
	URL  string
	Size int64
}

func (r ReleaseAsset) LogValue() any {
	return map[string]any{"name": r.Name, "size": r.Size}
}
