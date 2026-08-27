package dictionary

import (
	"reflect"
	"slices"
	"strings"
	"unicode"

	"github.com/eugene-static/pdfego/internal/components/parameter"
	"github.com/eugene-static/pdfego/internal/components/stream"
	"github.com/eugene-static/pdfego/internal/components/stream/bytes"
	"github.com/eugene-static/pdfego/internal/components/stream/primitives"
)

type Dictionary map[parameter.Name]*parameter.Parameter

func New(parameters ...*parameter.Parameter) *Dictionary {
	dict := make(Dictionary, len(parameters))

	for _, param := range parameters {
		dict[param.Name()] = param
	}

	return &dict
}

func Parse(body []byte) (*Dictionary, error) {
	dict := make(Dictionary)

	body = bytes.TrimSpace(body)
	body = bytes.TrimSuffix(body, []byte(primitives.DictionaryClose))
	bodyLen := len(body)

	for i := 0; i < bodyLen; {
		for i < bodyLen && body[i] != primitives.NamePrefix {
			i++
		}

		if i == bodyLen {
			break
		}

		startName := i + 1
		endName := startName

		for endName < bodyLen && (unicode.IsLetter(rune(body[endName])) || unicode.IsDigit(rune(body[endName])) || body[endName] == '_' || body[endName] == '.') {
			endName++
		}

		if endName == startName {
			i++

			continue
		}

		name := parameter.Name(body[startName:endName])

		i = endName

		for i < bodyLen && unicode.IsSpace(rune(body[i])) {
			i++
		}

		if i == bodyLen {
			dict[name] = &parameter.Parameter{}

			break
		}

		startValue := i
		endValue := startValue

		var (
			value stream.Appender
			err   error
		)

		switch body[startValue] {
		case primitives.HexadecimalStringOpen:
			if startValue+1 < bodyLen && body[startValue+1] == primitives.HexadecimalStringOpen {
				endValue = findParameterCloseToken(
					body,
					startValue,
					string(primitives.DictionaryOpen),
					string(primitives.DictionaryClose),
				)

				value, err = Parse(body[startValue:endValue])
				if err != nil {
					return nil, err
				}
			} else {
				endValue = findParameterCloseToken(
					body,
					startValue,
					string(primitives.HexadecimalStringOpen),
					string(primitives.HexadecimalStringClose),
				)

				value = parameter.TextHEX(body[startValue:endValue])
			}
		case primitives.ArrayOpen:
			endValue = findParameterCloseToken(
				body,
				startValue,
				string(primitives.ArrayOpen),
				string(primitives.ArrayClose),
			)

			value, err = parameter.ParseArray(body[startValue:endValue])
			if err != nil {
				return nil, err
			}
		case primitives.LiteralStringOpen:
			endValue = findParameterCloseToken(
				body,
				startValue,
				string(primitives.LiteralStringOpen),
				string(primitives.LiteralStringClose),
			)

			value = parameter.Text(body[startValue:endValue])
		default:
			endValue++

			for endValue < bodyLen && body[endValue] != primitives.NamePrefix {
				endValue++
			}

			value, err = parameter.ParseString(body[startValue:endValue])
			if err != nil {
				return nil, err
			}
		}

		dict[name] = parameter.New(name, value)

		i = endValue
	}

	return &dict, nil
}

func (dict Dictionary) Encode(value any) {
	v := reflect.ValueOf(value)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	typ := v.Type()
	numField := typ.NumField()

	for i := range numField {
		field := typ.Field(i)

		tag := field.Tag.Get("pdf")
		if tag == "" || tag == "-" {
			continue
		}

		parts := strings.Split(tag, ",")

		name := parameter.Name(parts[0])
		if name == "" {
			name = parameter.Name(field.Name)
		}

		skipEmpty := slices.Contains(parts[1:], "omitempty")

		val, ok := v.Field(i).Interface().(stream.Appender)
		if !ok {
			//TODO:
			continue
		}

		if isEmptyValue(val) && skipEmpty {
			continue
		}

		dict[name] = parameter.New(name, val)
	}
}

func (dict Dictionary) Get(name parameter.Name) (*parameter.Parameter, bool) {
	param, ok := dict[name]

	return param, ok
}

func (dict Dictionary) Set(name parameter.Name, value stream.Appender) {
	dict[name] = parameter.New(name, value)
}

func (dict Dictionary) Append(dst []byte) []byte {
	dst = append(dst, primitives.DictionaryOpen...)

	i := 0

	for _, param := range dict {
		if i > 0 {
			dst = bytes.AppendSpace(dst)
		}

		dst = param.Append(dst)

		i++
	}

	dst = append(dst, primitives.DictionaryClose...)

	return dst
}

func findParameterCloseToken(body []byte, start int, open, close string) int {
	bodyLen := len(body)
	openLen := len(open)
	closeLen := len(close)
	level := 0

	for i := start; i < bodyLen; {
		if i+openLen < bodyLen && string(body[i:i+openLen]) == open {
			level++
			i += openLen

			continue
		}

		if i+closeLen < bodyLen && string(body[i:i+closeLen]) == close {
			level--
			if level == 0 {
				return i + closeLen
			}

			i += closeLen

			continue
		}

		i++
	}

	return bodyLen
}

func isEmptyValue(val stream.Appender) bool {
	switch v := val.(type) {
	case parameter.Reference:
		return v == 0
	case *parameter.Reference:
		return v == nil
	case parameter.Name:
		return v == ""
	case parameter.References:
		return len(v) == 0
	case parameter.Names:
		return len(v) == 0
	case *Dictionary:
		return v == nil || len(*v) == 0
	case Dictionary:
		return v == nil || len(v) == 0
	default:
		return false
	}
}
