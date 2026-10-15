package webservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/updatefailure"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/dto"
	"github.com/brienze1/xenon/pkg/errs"
	httpClient "github.com/brienze1/xenon/pkg/http"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	updateCheckTimeout    = 15 * time.Second
	updateDownloadTimeout = 10 * time.Minute
	updateProgressEvery   = 250 * time.Millisecond
	updateFolder          = "updates"
	updatePartSuffix      = ".part"
	updateFolderMode      = 0o700
	updateBufferSize      = 32 << 10
)

type updateWebservice struct {
	newClient httpClient.ClientFactory
}

func NewUpdateWebservice(newClient httpClient.ClientFactory) adapter.UpdateWebservice {
	return &updateWebservice{newClient: newClient}
}

func (u *updateWebservice) FindLatestRelease(ctx context.Context) (*entity.Release, error) {
	logger.Debug(ctx, "Started")

	response, err := u.get(ctx, properties.Properties().Update.ApiURL+"/releases/latest", gitHubAccept, updateCheckTimeout)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, errs.ServerError("GitHub answered the release check with "+response.Status, "UPD-04502", http.StatusBadGateway)
	}
	var release dto.GitHubRelease
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return nil, errs.ServerError("GitHub answered the release check with an unreadable release", "UPD-05502", http.StatusBadGateway)
	}
	found, err := release.ToEntity()
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", found)
	return found, nil
}

func (u *updateWebservice) FindReleaseChecksums(ctx context.Context, asset entity.ReleaseAsset) (entity.ReleaseChecksums, error) {
	logger.Debug(ctx, "Started", asset)

	response, err := u.get(ctx, asset.URL, "", updateCheckTimeout)
	if err != nil {
		return nil, entity.NewUpdateFailure(updatefailure.DownloadFailed, err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, entity.NewUpdateFailure(updatefailure.DownloadFailed, errs.ServerError("The release checksums "+asset.Name+" answered "+response.Status, "UPD-09502", http.StatusBadGateway))
	}
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, entity.NewUpdateFailure(updatefailure.DownloadFailed, errs.ServiceUnavailableError(err, "UPD-10503", "The release checksums "+asset.Name+" could not be read"))
	}
	checksums := dto.ReleaseChecksums(content).ToEntity()

	logger.Debug(ctx, "Finished", asset, map[string]any{"checksums": checksums})
	return checksums, nil
}

func (u *updateWebservice) DownloadUpdatePackage(ctx context.Context, asset entity.ReleaseAsset, checksums entity.ReleaseChecksums, progress func(received, total int64)) (string, error) {
	logger.Debug(ctx, "Started", asset, map[string]any{"checksums": checksums})

	folder := filepath.Join(properties.Properties().Application.DataDir, updateFolder)
	if err := os.MkdirAll(folder, updateFolderMode); err != nil {
		return "", entity.NewUpdateFailure(updatefailure.DownloadFailed, errs.InternalServerError(err, "UPD-11500", "The update download folder could not be created"))
	}
	target := filepath.Join(folder, filepath.Base(asset.Name))
	part := target + updatePartSuffix
	digest, err := u.fetch(ctx, asset, part, progress)
	if err != nil {
		_ = os.Remove(part)
		return "", entity.NewUpdateFailure(updatefailure.DownloadFailed, err)
	}
	if !checksums.Matches(asset.Name, digest) {
		_ = os.Remove(part)
		return "", entity.NewUpdateFailure(updatefailure.Unverified, errs.UnprocessableEntityError("The update package "+asset.Name+" does not match its checksum", "UPD-13422"))
	}
	if err := os.Rename(part, target); err != nil {
		_ = os.Remove(part)
		return "", entity.NewUpdateFailure(updatefailure.DownloadFailed, errs.InternalServerError(err, "UPD-14500", "The update package "+asset.Name+" could not be moved into place"))
	}

	logger.Debug(ctx, "Finished", asset, map[string]any{"checksums": checksums, "path": target})
	return target, nil
}

func (u *updateWebservice) fetch(ctx context.Context, asset entity.ReleaseAsset, part string, progress func(received, total int64)) (string, error) {
	response, err := u.get(ctx, asset.URL, "", updateDownloadTimeout)
	if err != nil {
		return "", err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return "", errs.ServerError("The update package "+asset.Name+" answered "+response.Status, "UPD-20502", http.StatusBadGateway)
	}
	file, err := os.Create(part)
	if err != nil {
		return "", errs.InternalServerError(err, "UPD-21500", "The update package "+asset.Name+" could not be created")
	}
	defer func() { _ = file.Close() }()
	total := asset.ExpectedSize(response.ContentLength)
	hash := sha256.New()
	received, err := copyWithProgress(io.MultiWriter(file, hash), response.Body, total, progress)
	if err != nil {
		return "", err
	}
	progress(received, total)
	if err := file.Close(); err != nil {
		return "", errs.InternalServerError(err, "UPD-24500", "The update package "+asset.Name+" could not be saved")
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func copyWithProgress(target io.Writer, source io.Reader, total int64, progress func(received, total int64)) (int64, error) {
	buffer := make([]byte, updateBufferSize)
	var received int64
	reported := time.Now()
	progress(0, total)
	for {
		read, err := source.Read(buffer)
		if read > 0 {
			if _, writeErr := target.Write(buffer[:read]); writeErr != nil {
				return received, errs.InternalServerError(writeErr, "UPD-22500", "The update package could not be written")
			}
			received += int64(read)
			if time.Since(reported) >= updateProgressEvery {
				progress(received, total)
				reported = time.Now()
			}
		}
		if errors.Is(err, io.EOF) {
			return received, nil
		}
		if err != nil {
			return received, errs.ServiceUnavailableError(err, "UPD-23503", "The update download was interrupted")
		}
	}
}

func (u *updateWebservice) get(ctx context.Context, target, accept string, timeout time.Duration) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, errs.InternalServerError(err, "UPD-18500", "The update request could not be built")
	}
	if accept != "" {
		request.Header.Set("Accept", accept)
	}
	response, err := u.newClient.Build(httpClient.Options{Timeout: timeout, FollowRedirects: true, VerifySsl: true}).Do(ctx, request)
	if err != nil {
		return nil, errs.ServiceUnavailableError(err, "UPD-19503", "The update server could not be reached")
	}
	return response, nil
}
