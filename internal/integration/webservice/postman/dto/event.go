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

const javaScriptType = "text/javascript"

func NewEvents(own entity.Scripts, inherited []entity.Scripts, passthrough any) []Event {
	metadata := asMap(passthrough)
	events := make([]Event, 0)
	for _, scripts := range inherited {
		events = appendEvent(events, scripts, scriptphase.PreRequest, nil)
		events = appendEvent(events, scripts, scriptphase.PostResponse, nil)
	}
	events = appendEvent(events, own, scriptphase.PreRequest, metadata[preRequestListen])
	return appendEvent(events, own, scriptphase.PostResponse, metadata[testListen])
}

func appendEvent(events []Event, scripts entity.Scripts, phase enums.ScriptPhase, metadata any) []Event {
	script := scripts.Text(phase)
	if script == "" && metadata == nil {
		return events
	}
	listen := preRequestListen
	if phase == scriptphase.PostResponse {
		listen = testListen
	}
	details := asMap(metadata)
	if metadata == nil {
		details = map[string]any{"type": javaScriptType}
	}
	return append(events, Event{Listen: listen, Disabled: scripts.IsDisabled(phase), Script: Script{Exec: scriptLines(script), Raw: details}})
}

func (e Event) MarshalJSON() ([]byte, error) {
	event := newObject().set("listen", e.Listen).set("script", e.Script).setIf(e.Disabled, "disabled", true)
	return event.MarshalJSON()
}
