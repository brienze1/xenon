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
