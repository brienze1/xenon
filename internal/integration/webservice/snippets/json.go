package snippets

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
)

type jsonKind int

const (
	jsonNull jsonKind = iota
	jsonBool
	jsonNumber
	jsonString
	jsonArray
	jsonObject
)

const maxJsonIndent = 10

type jsonMember struct {
	key   string
	value jsonValue
}

type jsonValue struct {
	text    string
	items   []jsonValue
	members []jsonMember
	number  float64
	kind    jsonKind
	boolean bool
}

type jsonLiterals struct {
	null, yes, no string
}

var javascriptLiterals = jsonLiterals{null: "null", yes: "true", no: "false"}

var pythonLiterals = jsonLiterals{null: "None", yes: "True", no: "False"}

func parseJson(text string) (jsonValue, bool) {
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	value, err := decodeJson(decoder)
	if err != nil {
		return jsonValue{}, false
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return jsonValue{}, false
	}
	return value, true
}

func decodeJson(decoder *json.Decoder) (jsonValue, error) {
	token, err := decoder.Token()
	if err != nil {
		return jsonValue{}, err
	}
	switch typed := token.(type) {
	case json.Delim:
		if typed == '[' {
			return decodeJsonArray(decoder)
		}
		return decodeJsonObject(decoder)
	case string:
		return jsonValue{kind: jsonString, text: typed}, nil
	case json.Number:
		number, _ := strconv.ParseFloat(typed.String(), 64)
		return jsonValue{kind: jsonNumber, number: number}, nil
	case bool:
		return jsonValue{kind: jsonBool, boolean: typed}, nil
	default:
		return jsonValue{kind: jsonNull}, nil
	}
}

func decodeJsonArray(decoder *json.Decoder) (jsonValue, error) {
	array := jsonValue{kind: jsonArray, items: []jsonValue{}}
	for decoder.More() {
		item, err := decodeJson(decoder)
		if err != nil {
			return jsonValue{}, err
		}
		array.items = append(array.items, item)
	}
	_, err := decoder.Token()
	return array, err
}

func decodeJsonObject(decoder *json.Decoder) (jsonValue, error) {
	object := jsonValue{kind: jsonObject, members: []jsonMember{}}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return jsonValue{}, err
		}
		key, _ := token.(string)
		value, err := decodeJson(decoder)
		if err != nil {
			return jsonValue{}, err
		}
		object.members = setMember(object.members, key, value)
	}
	_, err := decoder.Token()
	return object, err
}

func setMember(members []jsonMember, key string, value jsonValue) []jsonMember {
	for index := range members {
		if members[index].key == key {
			members[index].value = value
			return members
		}
	}
	return append(members, jsonMember{key: key, value: value})
}

func (v jsonValue) orderedMembers() []jsonMember {
	indexed := make([]jsonMember, 0, len(v.members))
	named := make([]jsonMember, 0, len(v.members))
	for _, member := range v.members {
		if isArrayIndex(member.key) {
			indexed = insertByIndex(indexed, member)
			continue
		}
		named = append(named, member)
	}
	return append(indexed, named...)
}

func isArrayIndex(key string) bool {
	index, err := strconv.ParseUint(key, 10, 32)
	return err == nil && index < math.MaxUint32 && strconv.FormatUint(index, 10) == key
}

func insertByIndex(members []jsonMember, member jsonMember) []jsonMember {
	index, _ := strconv.ParseUint(member.key, 10, 32)
	position := len(members)
	for at, existing := range members {
		if other, _ := strconv.ParseUint(existing.key, 10, 32); other > index {
			position = at
			break
		}
	}
	return append(members[:position], append([]jsonMember{member}, members[position:]...)...)
}

func stringifyJson(value jsonValue, indent int, literals jsonLiterals) string {
	var builder strings.Builder
	writeJson(&builder, value, strings.Repeat(" ", min(max(indent, 0), maxJsonIndent)), "", literals)
	return builder.String()
}

func writeJson(builder *strings.Builder, value jsonValue, gap, current string, literals jsonLiterals) {
	switch value.kind {
	case jsonNull:
		builder.WriteString(literals.null)
	case jsonBool:
		builder.WriteString(map[bool]string{true: literals.yes, false: literals.no}[value.boolean])
	case jsonNumber:
		builder.WriteString(javascriptNumber(value.number))
	case jsonString:
		builder.WriteString(quoteJson(value.text))
	case jsonArray:
		writeJsonList(builder, "[", "]", len(value.items), gap, current, func(index int, inner string) {
			writeJson(builder, value.items[index], gap, inner, literals)
		})
	default:
		members := value.orderedMembers()
		separator := ":"
		if gap != "" {
			separator = ": "
		}
		writeJsonList(builder, "{", "}", len(members), gap, current, func(index int, inner string) {
			builder.WriteString(quoteJson(members[index].key))
			builder.WriteString(separator)
			writeJson(builder, members[index].value, gap, inner, literals)
		})
	}
}

func writeJsonList(builder *strings.Builder, open, closing string, count int, gap, current string, write func(index int, inner string)) {
	builder.WriteString(open)
	if count == 0 {
		builder.WriteString(closing)
		return
	}
	inner := current + gap
	for index := 0; index < count; index++ {
		if index > 0 {
			builder.WriteString(",")
		}
		if gap != "" {
			builder.WriteString("\n" + inner)
		}
		write(index, inner)
	}
	if gap != "" {
		builder.WriteString("\n" + current)
	}
	builder.WriteString(closing)
}

func quoteJson(text string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for _, character := range text {
		switch character {
		case '"':
			builder.WriteString(`\"`)
		case '\\':
			builder.WriteString(`\\`)
		case '\b':
			builder.WriteString(`\b`)
		case '\f':
			builder.WriteString(`\f`)
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		default:
			if character < 0x20 {
				builder.WriteString(`\u00` + strconv.FormatInt(int64(character)>>4, 16) + strconv.FormatInt(int64(character)&0xf, 16))
				continue
			}
			builder.WriteRune(character)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}

func javascriptNumber(number float64) string {
	if math.IsInf(number, 0) || math.IsNaN(number) {
		return "null"
	}
	if number == 0 {
		return "0"
	}
	sign := ""
	if number < 0 {
		sign, number = "-", -number
	}
	scientific := strconv.FormatFloat(number, 'e', -1, 64)
	mantissa, exponentText, _ := strings.Cut(scientific, "e")
	digits := strings.Replace(mantissa, ".", "", 1)
	exponent, _ := strconv.Atoi(exponentText)
	point := exponent + 1
	count := len(digits)
	switch {
	case count <= point && point <= 21:
		return sign + digits + strings.Repeat("0", point-count)
	case 0 < point && point <= 21:
		return sign + digits[:point] + "." + digits[point:]
	case -6 < point && point <= 0:
		return sign + "0." + strings.Repeat("0", -point) + digits
	}
	fraction := ""
	if count > 1 {
		fraction = "." + digits[1:]
	}
	exponentSign := "+"
	if point-1 < 0 {
		exponentSign = "-"
	}
	return sign + digits[:1] + fraction + "e" + exponentSign + strconv.Itoa(abs(point-1))
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
