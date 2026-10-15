package dto

import (
	"net/http"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

type GitHubRelease struct {
	TagName    string               `json:"tag_name"`
	HTMLURL    string               `json:"html_url"`
	Assets     []GitHubReleaseAsset `json:"assets"`
	Draft      bool                 `json:"draft"`
	Prerelease bool                 `json:"prerelease"`
}

type GitHubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

func (g GitHubRelease) ToEntity() (*entity.Release, error) {
	if g.Draft || g.Prerelease {
		return nil, errs.ServerError("GitHub answered with a release that is not published", "UPD-03502", http.StatusBadGateway)
	}
	version, err := entity.ParseVersion(g.TagName)
	if err != nil {
		return nil, err
	}
	assets := make([]entity.ReleaseAsset, 0, len(g.Assets))
	for _, asset := range g.Assets {
		assets = append(assets, entity.ReleaseAsset{Name: asset.Name, URL: asset.BrowserDownloadURL, Size: asset.Size})
	}
	return &entity.Release{Version: version, PageURL: g.HTMLURL, Assets: assets}, nil
}
