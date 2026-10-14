package entity_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	redactionToken    = "bearer-token-never-logged"
	redactionPassword = "basic-password-never-logged"
	redactionHeader   = "header-value-never-logged"
	redactionBody     = "body-content-never-logged"
)

func captureEntityLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	buffer := &bytes.Buffer{}
	logger.SetOutput(buffer)
	t.Cleanup(func() { logger.SetOutput(os.Stdout) })
	return buffer
}

func secretAuth(authType string) entity.Auth {
	auth := entity.DefaultAuth()
	auth.Type = authtype.GetAuthType(authType)
	auth.Bearer = entity.AuthBearer{Token: redactionToken}
	auth.Basic = entity.AuthBasic{Username: "user", Password: redactionPassword}
	return auth
}

func secretBody() entity.Body {
	body := entity.DefaultBody()
	body.Mode = bodymode.Raw
	body.Raw = entity.NewBodyRaw("json", redactionBody)
	return body
}

func TestEffectiveRequestLogNeverContainsSecrets(t *testing.T) {
	logs := captureEntityLogs(t)
	request := entity.EffectiveRequest{
		Auth:        secretAuth("bearer"),
		Body:        secretBody(),
		Method:      "POST",
		URL:         "https://api.example.test/orders",
		Headers:     []entity.KeyValue{entity.NewKeyValue("X-Api-Key", redactionHeader)},
		AutoHeaders: []entity.AutoHeader{entity.NewAutoHeader(entity.HeaderAuthorization, "Bearer "+redactionToken)},
	}

	logger.Info(context.Background(), "Started", request, map[string]any{"request": request})

	assertRedacted(t, logs.String(), "https://api.example.test/orders")
}

func TestHistoryEntryLogNeverContainsSecrets(t *testing.T) {
	logs := captureEntityLogs(t)
	request := entity.NewRequest("req-1", "Orders")
	request.URL = "https://api.example.test/history"
	request.Auth = secretAuth("basic")
	request.Body = secretBody()
	request.Headers = []entity.KeyValue{entity.NewKeyValue("Cookie", redactionHeader)}
	response := entity.NewResponse(200, "OK", "HTTP/1.1", []entity.KeyValue{entity.NewKeyValue("Set-Cookie", redactionHeader)}, []byte(redactionBody), 10, entity.Timings{})
	execution := entity.NewSendExecution("send-1")
	execution.Start(request, entity.ScriptScopes{Local: entity.NewScriptScope("", []entity.ScriptVariable{entity.NewScriptVariable("token", redactionHeader)})}, map[string]string{"key": redactionHeader})
	execution.RecordResponse(entity.EffectiveRequest{URL: "https://api.example.test/history", Headers: request.Headers}, response)
	entry := entity.NewHistoryEntry("entry-1", time.Now(), request, execution, entity.UserSendOrigin())

	logger.Info(context.Background(), "Finished", entry, entity.HistoryEntries{entry}, map[string]any{"result": entry, "response": response, "execution": &execution})

	assertRedacted(t, logs.String(), "https://api.example.test/history")
}

func assertRedacted(t *testing.T, output, identifier string) {
	t.Helper()
	if !strings.Contains(output, identifier) {
		t.Fatalf("the log must still identify the request:\n%s", output)
	}
	for _, secret := range []string{redactionToken, redactionPassword, redactionHeader, redactionBody} {
		if strings.Contains(output, secret) {
			t.Fatalf("secret %q reached the log output:\n%s", secret, output)
		}
	}
}
