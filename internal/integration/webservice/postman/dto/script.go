package dto

import "encoding/json"

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

func (s Script) MarshalJSON() ([]byte, error) {
	script := newObject()
	if id, found := s.Raw["id"]; found {
		script.set("id", id)
	}
	script.setPresent("type", s.Raw["type"]).set("exec", s.Exec)
	return script.merged(s.Raw).MarshalJSON()
}
