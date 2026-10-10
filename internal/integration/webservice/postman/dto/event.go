package dto

import (
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
)

const (
	preRequestListen = "prerequest"
	testListen       = "test"
)

type Script struct {
	Raw  map[string]any `json:"-"`
	Exec any            `json:"exec"`
}

func (s *Script) UnmarshalJSON(data []byte) error {
	type plain Script
	if err := json.Unmarshal(data, (*plain)(s)); err != nil {
		return err
	}
	s.Raw = decodeRaw(data)
	return nil
}

type Event struct {
	Script   Script         `json:"script"`
	Raw      map[string]any `json:"-"`
	Listen   string         `json:"listen"`
	Disabled bool           `json:"disabled"`
}

func (e *Event) UnmarshalJSON(data []byte) error {
	type plain Event
	if err := json.Unmarshal(data, (*plain)(e)); err != nil {
		return err
	}
	e.Raw = decodeRaw(data)
	return nil
}

func (e *Event) phase() enums.ScriptPhase {
	switch e.Listen {
	case preRequestListen:
		return scriptphase.PreRequest
	case testListen:
		return scriptphase.PostResponse
	default:
		return scriptphase.Unknown
	}
}

func scriptsToEntity(events []Event) entity.Scripts {
	scripts := entity.Scripts{}
	for index := range events {
		event := events[index]
		switch event.phase() {
		case scriptphase.PreRequest:
			scripts.PreRequest = lines(event.Script.Exec)
		case scriptphase.PostResponse:
			scripts.PostResponse = lines(event.Script.Exec)
		default:
			continue
		}
		if event.Disabled {
			scripts.Disabled = append(scripts.Disabled, event.phase())
		}
	}
	return scripts
}

func eventsPassthrough(events []Event) map[string]any {
	metadata := map[string]any{}
	for index := range events {
		event := events[index]
		details := without(event.Raw, "listen", "script", "disabled")
		for key, value := range without(event.Script.Raw, "exec") {
			details[key] = value
		}
		metadata = withEntry(metadata, event.Listen, details)
	}
	return metadata
}
