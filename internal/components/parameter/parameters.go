package parameter

import (
	"regexp"

	"github.com/eugene-static/pdfego/internal/components/stream"
	"github.com/eugene-static/pdfego/internal/components/stream/bytes"
	"github.com/eugene-static/pdfego/internal/components/stream/primitives"
)

var (
	regexReference     = regexp.MustCompile(`(\d+)\s\d\sR`)
	regexNumber        = regexp.MustCompile(`(-?\d+\.?\d*)`)
	regexNumberInteger = regexp.MustCompile(`(\d+)`)
	regexName          = regexp.MustCompile(`/(\S+)`)
	regexBool          = regexp.MustCompile(`(true|false)`)

	regexReferenceArray = regexp.MustCompile(`\[\s*(?:\s*\d+\s+\d+\s+R\s*)+]`)
	regexNumberArray    = regexp.MustCompile(`\[\s*(?:\d+\.?\d*\s*)+]`)
	regexNameArray      = regexp.MustCompile(`\[\s*(?:/\S+\s*)+]`)
)

type Parameter struct {
	name  Name
	value stream.Appender
}

func New(name Name, value stream.Appender) *Parameter {
	return &Parameter{
		name:  name,
		value: value,
	}
}

func (param *Parameter) Name() Name {
	return param.name
}

func (param *Parameter) Value() stream.Appender {
	return param.value
}

func (param *Parameter) Append(dst []byte) []byte {
	return bytes.NewWriter(dst).
		Write(param.name).SP().
		Write(param.value).Bytes()
}

func ParseArray(value []byte) (stream.Appender, error) {
	value = bytes.TrimSpace(value)

	switch {
	case regexReferenceArray.Match(value):
		submatches := regexReference.FindAllSubmatch(value, -1)

		refArray, err := parseReferenceArray(submatches)
		if err != nil {
			return nil, err
		}

		return &refArray, nil
	case regexNumberArray.Match(value):
		submatches := regexNumber.FindAllSubmatch(value, -1)

		numberArray, err := parseNumberArray(submatches)
		if err != nil {
			return nil, err
		}

		return &numberArray, nil
	case regexNameArray.Match(value):
		submatches := regexName.FindAllSubmatch(value, -1)

		nameArray := parseNameArray(submatches)

		return &nameArray, nil
	default:
		return Raw(value), nil
	}
}

func ParseString(value []byte) (stream.Appender, error) {
	value = bytes.TrimSpace(value)

	submatch := regexReference.FindSubmatch(value)
	if len(submatch) != 0 {
		ref, err := parseReference(submatch)
		if err != nil {
			return nil, err
		}

		return ref, nil
	}

	submatch = regexNumber.FindSubmatch(value)
	if len(submatch) != 0 {
		number, err := parseNumber(submatch)
		if err != nil {
			return nil, err
		}

		return number, nil
	}

	submatch = regexNumberInteger.FindSubmatch(value)
	if len(submatch) != 0 {
		number, err := parseInteger(submatch)
		if err != nil {
			return nil, err
		}

		return number, nil
	}

	submatch = regexName.FindSubmatch(value)
	if len(submatch) != 0 {
		name := parseName(submatch)

		return name, nil
	}

	submatch = regexBool.FindSubmatch(value)
	if len(submatch) != 0 {
		boolean := parseBool(submatch)

		return boolean, nil
	}

	return Raw(value), nil
}

type Raw []byte

func (param Raw) Append(dst []byte) []byte {
	return append(dst, param...)
}

type Text []byte

func (param Text) Append(dst []byte) []byte {
	dst = append(dst, primitives.LiteralStringOpen)
	dst = append(dst, param...)
	dst = append(dst, primitives.LiteralStringClose)

	return dst
}

type TextHEX []byte

func (param TextHEX) Append(dst []byte) []byte {
	return append(dst, param...)
}

type Reference int

func parseReference(submatch [][]byte) (Reference, error) {
	value, err := bytes.ParseInt(submatch[1])
	if err != nil {
		return 0, err
	}

	return Reference(value), nil
}

func (param Reference) Append(dst []byte) []byte {
	if param == 0 {
		return append(dst, primitives.Null...)
	}

	dst = bytes.AppendInt(dst, int64(param))
	dst = append(dst, " 0 R"...)

	return dst
}

type Number float64

func parseNumber(submatch [][]byte) (Number, error) {
	value, err := bytes.ParseFloat(submatch[1])
	if err != nil {
		return 0, err
	}

	return Number(value), nil
}

func (param Number) Append(dst []byte) []byte {
	return bytes.AppendFloat(dst, float64(param))
}

func (param Number) hasZeroDecimalPart() bool {
	return param == Number(int(param))
}

type Integer int

func parseInteger(submatch [][]byte) (Integer, error) {
	value, err := bytes.ParseInt(submatch[1])
	if err != nil {
		return 0, err
	}

	return Integer(value), nil
}

func (param Integer) Append(dst []byte) []byte {
	return bytes.AppendInt(dst, int64(param))
}

type Name string

func parseName(submatch [][]byte) Name {
	return Name(submatch[1])
}

func (param Name) Append(dst []byte) []byte {
	dst = append(dst, primitives.NamePrefix)

	return append(dst, param...)
}

type Bool bool

func parseBool(submatch [][]byte) Bool {
	value := primitives.Keyword(submatch[1])

	return value == primitives.True
}

func (p Bool) Append(dst []byte) []byte {
	return bytes.AppendBool(dst, bool(p))
}

type References []Reference

func parseReferenceArray(submatches [][][]byte) (References, error) {
	values := make([]Reference, 0, len(submatches))

	for _, submatch := range submatches {
		value, err := bytes.ParseInt(submatch[1])
		if err != nil {
			return nil, err
		}

		values = append(values, Reference(value))
	}

	return values, nil
}

func (param References) Append(dst []byte) []byte {
	dst = append(dst, primitives.ArrayOpen)

	for i, ref := range param {
		if i > 0 {
			dst = bytes.AppendSpace(dst)
		}

		dst = ref.Append(dst)
	}

	dst = append(dst, primitives.ArrayClose)

	return dst
}

type ReferencesPtrs []*Reference

func (param ReferencesPtrs) Append(dst []byte) []byte {
	dst = append(dst, primitives.ArrayOpen)

	for i, ref := range param {
		if ref == nil || *ref == 0 {
			continue
		}

		if i > 0 {
			dst = bytes.AppendSpace(dst)
		}

		dst = ref.Append(dst)
	}

	dst = append(dst, primitives.ArrayClose)

	return dst
}

type Numbers []Number

func NewNumberArray[T int | float64](vals ...T) Numbers {
	arr := make(Numbers, 0, len(vals))

	for _, val := range vals {
		arr = append(arr, Number(val))
	}

	return arr
}

func parseNumberArray(submatches [][][]byte) (Numbers, error) {
	values := make([]Number, 0, len(submatches))

	for _, submatch := range submatches {
		value, err := bytes.ParseFloat(submatch[1])
		if err != nil {
			return Numbers{}, err
		}

		values = append(values, Number(value))
	}

	return values, nil
}

func (param Numbers) Append(dst []byte) []byte {
	dst = append(dst, primitives.ArrayOpen)

	for i, val := range param {
		if i > 0 {
			dst = bytes.AppendSpace(dst)
		}

		if val.hasZeroDecimalPart() {
			dst = Integer(val).Append(dst)

			continue
		}

		dst = val.Append(dst)
	}

	dst = append(dst, primitives.ArrayClose)

	return dst
}

type Names []Name

func parseNameArray(submatches [][][]byte) Names {
	values := make([]Name, 0, len(submatches))

	for _, submatch := range submatches {
		values = append(values, Name(submatch[1]))
	}

	return values
}

func (param Names) Append(dst []byte) []byte {
	dst = append(dst, primitives.ArrayOpen)

	for i, val := range param {
		if i > 0 {
			dst = bytes.AppendSpace(dst)
		}

		dst = val.Append(dst)
	}

	dst = append(dst, primitives.ArrayClose)

	return dst
}

type GlyphsWidthTable [][2]uint64

func (param GlyphsWidthTable) Append(dst []byte) []byte {
	dst = append(dst, primitives.ArrayOpen)

	prev := uint64(0)

	for _, gl := range param {
		index := gl[0]
		advance := gl[1]

		if index == prev+1 {
			dst = bytes.AppendSpace(dst)
			dst = bytes.AppendUint(dst, advance)

			continue
		}

		if prev > 0 {
			dst = append(dst, primitives.ArrayClose, ' ')
		}

		dst = bytes.AppendUint(dst, index)
		dst = append(dst, ' ', primitives.ArrayOpen)
		dst = bytes.AppendUint(dst, advance)

		prev = index
	}

	dst = append(dst, primitives.ArrayClose)
	dst = append(dst, primitives.ArrayClose)

	return dst
}
