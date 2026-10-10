package dto

import "encoding/json"

type Description struct {
	Raw     map[string]any
	Content string
}

func (d *Description) UnmarshalJSON(data []byte) error {
	var content string
	if err := json.Unmarshal(data, &content); err == nil {
		d.Content = content
		return nil
	}
	d.Raw = decodeRaw(data)
	d.Content = text(d.Raw["content"])
	return nil
}

func (d Description) Passthrough() map[string]any {
	return without(d.Raw, "content")
}
