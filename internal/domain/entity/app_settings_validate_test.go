package entity

import (
	"testing"

	"github.com/brienze1/xenon/pkg/errs"
)

func TestAppSettingsValidateReturnsTheFirstFailingCode(t *testing.T) {
	if err := DefaultAppSettings().Validate(); err != nil {
		t.Fatalf("valid settings: %v", err)
	}
	cases := map[string]func(*AppSettings){
		"SCR-01400": func(s *AppSettings) { s.ScriptTimeoutMs, s.HistoryUserCap = 999, 99 },
		"SET-02400": func(s *AppSettings) { s.HistoryUserCap, s.HistoryAICap = 20001, 5 },
		"SET-03400": func(s *AppSettings) { s.HistoryAICap = 99 },
		"SET-04400": func(s *AppSettings) { s.AIAccess.Port = 1023 },
	}
	for code, breakSettings := range cases {
		settings := DefaultAppSettings()
		breakSettings(&settings)
		failure, ok := settings.Validate().(errs.BaseError)
		if !ok || failure.Code() != code {
			t.Errorf("want %s, got %v", code, failure)
		}
	}
}
