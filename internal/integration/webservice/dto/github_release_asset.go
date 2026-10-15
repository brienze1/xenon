package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type GitHubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

func (g GitHubReleaseAsset) ToEntity() entity.ReleaseAsset {
	return entity.ReleaseAsset{Name: g.Name, URL: g.BrowserDownloadURL, Size: g.Size}
}
