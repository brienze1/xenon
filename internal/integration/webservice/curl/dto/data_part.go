package dto

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

var formKey = regexp.MustCompile(`^[A-Za-z0-9_.\-\[\]%]+$`)

type DataPart struct {
	Value      string
	Literal    bool
	UrlEncoded bool
}

func NewDataPart(value string, literal bool) DataPart {
	return DataPart{Value: value, Literal: literal}
}

func NewUrlEncodedPart(value string) DataPart {
	return DataPart{Value: value, UrlEncoded: true}
}

func (d DataPart) IsFile() bool {
	return !d.Literal && !d.UrlEncoded && strings.HasPrefix(d.Value, filePrefix)
}

func (d DataPart) IsFormLike() bool {
	if d.UrlEncoded {
		return true
	}
	for _, pair := range strings.Split(d.Value, "&") {
		key, _, found := strings.Cut(pair, "=")
		if !found || !formKey.MatchString(key) {
			return false
		}
	}
	return true
}

func (d DataPart) Rows() []entity.KeyValue {
	if d.UrlEncoded {
		key, value := d.urlEncodedPair()
		return []entity.KeyValue{entity.NewKeyValue(key, value)}
	}
	rows := make([]entity.KeyValue, 0)
	for _, pair := range strings.Split(d.Value, "&") {
		key, value, _ := strings.Cut(pair, "=")
		rows = append(rows, entity.NewKeyValue(unescape(key), unescape(value)))
	}
	return rows
}

func (d DataPart) Query() string {
	if !d.UrlEncoded {
		return d.Value
	}
	key, value := d.urlEncodedPair()
	if key == "" {
		return queryEscape(value)
	}
	return key + "=" + queryEscape(value)
}

func (d DataPart) urlEncodedPair() (string, string) {
	key, value, found := strings.Cut(d.Value, "=")
	if !found {
		return "", d.Value
	}
	return key, value
}

func unescape(value string) string {
	if decoded, err := url.QueryUnescape(value); err == nil {
		return decoded
	}
	return value
}
