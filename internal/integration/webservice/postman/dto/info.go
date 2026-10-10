package dto

import "encoding/json"

type Info struct {
	Raw         map[string]any `json:"-"`
	Name        string         `json:"name"`
	Description Description    `json:"description"`
}

func (i *Info) UnmarshalJSON(data []byte) error {
	type plain Info
	if err := json.Unmarshal(data, (*plain)(i)); err != nil {
		return err
	}
	i.Raw = decodeRaw(data)
	return nil
}
