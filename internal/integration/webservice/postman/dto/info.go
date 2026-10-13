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

const collectionSchema = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"

func NewInfo(name, description string, id any, passthrough map[string]any) Info {
	fields := without(passthrough, "name", "description", "schema", "_postman_id")
	fields["_postman_id"] = id
	return Info{Name: name, Description: newDescription(description, passthrough["description"]), Raw: fields}
}

func (i Info) MarshalJSON() ([]byte, error) {
	info := newObject().set("_postman_id", i.Raw["_postman_id"]).set("name", i.Name).
		setPresent("description", i.Description.Value()).
		set("schema", collectionSchema)
	return info.merged(i.Raw).MarshalJSON()
}
