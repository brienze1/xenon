package webservice

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/dto"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const (
	awsHeaderPrefix      = "x-amz-"
	awsDateHeader        = "X-Amz-Date"
	awsTokenHeader       = "X-Amz-Security-Token"
	awsContentHeader     = "X-Amz-Content-Sha256"
	awsExpiresParam      = "X-Amz-Expires"
	awsPresignExpiration = 24 * time.Hour
)

type awsV4Webservice struct{}

func NewAwsV4Webservice() adapter.AwsV4Webservice {
	return &awsV4Webservice{}
}

func (a *awsV4Webservice) Sign(ctx context.Context, request entity.EffectiveRequest, auth entity.AwsV4Auth) (*entity.EffectiveRequest, error) {
	logger.Debug(ctx, "Started", map[string]any{"request": request, "auth": auth})

	request = withoutAwsSignature(request)
	payloadHash := usecase.ComputePayloadHash(request.Body, payload(request.Body))
	if auth.IsS3() && !auth.AddAuthDataToQuery {
		request = request.WithHeader(awsContentHeader, payloadHash)
	}
	signing, err := signingRequest(ctx, request)
	if err != nil {
		return nil, errs.BadRequestError("the request URL cannot be signed: "+err.Error(), "AWS-01400")
	}
	credentials := aws.Credentials{AccessKeyID: auth.AccessKey, SecretAccessKey: auth.SecretKey, SessionToken: auth.SessionToken}
	signer := v4.NewSigner(func(options *v4.SignerOptions) { options.DisableURIPathEscaping = auth.IsS3() })
	now := timeutils.Now()
	if auth.AddAuthDataToQuery {
		query := signing.URL.Query()
		query.Set(awsExpiresParam, strconv.Itoa(int(awsPresignExpiration.Seconds())))
		signing.URL.RawQuery = strings.ReplaceAll(query.Encode(), "+", "%20")
		signed, _, err := signer.PresignHTTP(ctx, credentials, signing, payloadHash, auth.ServiceOrDefault(), auth.RegionOrDefault(), now)
		if err != nil {
			return nil, errs.InternalServerError(err, "AWS-02500", "the request could not be presigned")
		}
		request = request.WithURL(signed)
		logger.Debug(ctx, "Finished", request)
		return &request, nil
	}
	if err := signer.SignHTTP(ctx, credentials, signing, payloadHash, auth.ServiceOrDefault(), auth.RegionOrDefault(), now); err != nil {
		return nil, errs.InternalServerError(err, "AWS-03500", "the request could not be signed")
	}
	for _, name := range []string{awsDateHeader, awsTokenHeader, entity.HeaderAuthorization} {
		if value := signing.Header.Get(name); value != "" {
			request = request.WithHeader(name, value)
		}
	}

	logger.Debug(ctx, "Finished", request)
	return &request, nil
}

func withoutAwsSignature(request entity.EffectiveRequest) entity.EffectiveRequest {
	request = request.WithoutHeader(entity.HeaderAuthorization)
	for _, header := range request.SendHeaders() {
		if strings.HasPrefix(strings.ToLower(header.Key), awsHeaderPrefix) {
			request = request.WithoutHeader(header.Key)
		}
	}
	for _, param := range entity.ParseRequestUrl(request.URL).QueryParams() {
		if key, err := url.QueryUnescape(param.Key); err == nil && strings.HasPrefix(strings.ToLower(key), awsHeaderPrefix) {
			request = request.WithoutQueryParam(key)
		}
	}
	return request
}

func signingRequest(ctx context.Context, request entity.EffectiveRequest) (*http.Request, error) {
	signing, err := http.NewRequestWithContext(ctx, request.Method, request.URL, http.NoBody)
	if err != nil {
		return nil, err
	}
	for _, header := range request.SendHeaders() {
		if strings.EqualFold(header.Key, entity.HeaderHost) {
			signing.Host = header.Value
			continue
		}
		signing.Header.Add(header.Key, header.Value)
	}
	return signing, nil
}

func payload(body entity.Body) []byte {
	files, sendErr := readFiles(body.FileRefs())
	if sendErr != nil {
		return nil
	}
	content := dto.NewRequestBody(body, files).Content
	if content == nil {
		return nil
	}
	data, _ := io.ReadAll(content)
	return bytes.Clone(data)
}
