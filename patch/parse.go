package patch

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"unicode"
)

type parameter struct {
	name  string
	value []byte
}

type dictionary map[string][]byte

func (dict dictionary) parameter(name string) (parameter, bool) {
	value, ok := dict[name]
	if !ok {
		return parameter{}, false
	}

	return parameter{name: name, value: value}, true
}

type trailer struct {
	index      int
	dictionary dictionary
}

func parseTrailer(body []byte) (*trailer, error) {
	trailerIndex := bytes.LastIndex(body, []byte("trailer"))
	if trailerIndex == -1 {
		err := errNotFound("trailer")

		return nil, err
	}

	trailerBody := body[trailerIndex:]

	startXrefIndex := bytes.Index(trailerBody, []byte("startxref"))
	if startXrefIndex == -1 {
		err := errNotFound("startxref")

		return nil, err
	}

	dict := parseDictionary(trailerBody[:startXrefIndex])

	return &trailer{
		index:      trailerIndex,
		dictionary: dict,
	}, nil
}

type xref struct {
	index   int
	offset  int
	count   int
	offsets []int
}

func parseXref(body []byte, trailerIndex int) (*xref, error) {
	startXrefIndex := bytes.LastIndex(body, []byte("startxref"))
	if startXrefIndex == -1 {
		err := errNotFound("startxref")

		return nil, err
	}

	eofIndex := bytes.LastIndex(body, []byte("%%EOF"))
	if eofIndex == -1 {
		err := errNotFound("%%EOF")

		return nil, err
	}

	xrefIndex, err := strconv.Atoi(string(bytes.TrimSpace(body[startXrefIndex:eofIndex])))
	if err != nil {
		return nil, err
	}

	xrefBody := body[xrefIndex:trailerIndex]

	lines := bytes.Split(xrefBody, []byte("\n"))
	if len(lines) < 2 {
		err = errors.New("таблица xref пуста")

		return nil, err
	}

	sum := bytes.Fields(lines[1])
	if len(sum) != 2 {
		err = errors.New("неверный формат суммы xref")

		return nil, err
	}

	offset := bytesToInt(sum[0])
	count := bytesToInt(sum[1])

	if count != len(lines)-2 {
		err = errors.New("таблица xref повреждена")

		return nil, err
	}

	objectsIndexes := make([]int, 0, count)

	for i := 2; i < len(lines); i++ {
		ref := bytes.Fields(lines[i])
		if len(ref) != 3 {
			err = errors.New("неверный формат таблицы xref")

			return nil, err
		}

		objectIndex := bytesToInt(ref[0])

		objectsIndexes = append(objectsIndexes, objectIndex)
	}

	return &xref{
		index:   xrefIndex,
		offset:  offset,
		count:   count,
		offsets: objectsIndexes,
	}, nil
}

func (x *xref) getOffset(objectNumber int) int {
	if objectNumber > len(x.offsets) {
		return -1
	}

	return x.offsets[objectNumber]
}

func parseDictionary(body []byte) dictionary {
	dict := make(map[string][]byte)

	body = bytes.TrimSpace(body)
	body = bytes.TrimSuffix(body, []byte(">>"))
	bodyLen := len(body)

	for i := 0; i < bodyLen; {
		for i < bodyLen && body[i] != '/' {
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

		name := string(body[startName:endName])

		i = endName

		for i < bodyLen && unicode.IsSpace(rune(body[i])) {
			i++
		}

		if i == bodyLen {
			dict[name] = nil

			break
		}

		startValue := i
		endValue := startValue

		switch body[startValue] {
		case '<':
			if startValue+1 < bodyLen && body[startValue+1] == '<' {
				endValue = findParameterCloseToken(body, startValue, "<<", ">>")
			}
		case '[':
			endValue = findParameterCloseToken(body, startValue, "[", "]")
		case '(':
			endValue = findParameterCloseToken(body, startValue, "(", ")")
		default:
			endValue++

			for endValue < bodyLen && body[endValue] != '/' {
				endValue++
			}
		}

		value := bytes.TrimSpace(body[startValue:endValue])

		dict[name] = value

		i = endValue
	}

	return dict
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

func parseStream(body []byte) (int, int) {
	startStreamIndex := bytes.Index(body, []byte("stream"))
	if startStreamIndex == -1 {
		return 0, 0
	}

	endStreamIndex := bytes.Index(body[startStreamIndex+1:], []byte("endstream"))
	if endStreamIndex == -1 {
		return 0, 0
	}

	return startStreamIndex + len("stream"), startStreamIndex + endStreamIndex
}

type parameterReferences struct {
	name          string
	objectNumbers []int
}

var regexReference = regexp.MustCompile(`(\d+)\s\d\sR`)

func (param *parameter) parseReferences() (parameterReferences, error) {
	submatches := regexReference.FindAllSubmatch(param.value, -1)
	if len(submatches) == 0 {
		err := errInvalidFormat(param.name)

		return parameterReferences{}, err
	}

	objectNumbers := make([]int, 0, len(submatches))

	for _, submatch := range submatches {
		objectNumber, err := strconv.Atoi(string(submatch[1]))
		if err != nil {
			return parameterReferences{}, err
		}

		objectNumbers = append(objectNumbers, objectNumber)
	}

	paramReference := parameterReferences{
		name:          param.name,
		objectNumbers: objectNumbers,
	}

	return paramReference, nil
}

type parameterReference struct {
	name         string
	objectNumber int
}

func (param *parameter) parseReference() (parameterReference, error) {
	paramRefs, err := param.parseReferences()
	if err != nil {
		return parameterReference{}, err
	}

	if len(paramRefs.objectNumbers) == 0 {
		err = errNotFound(param.name + " reference")
	}

	return parameterReference{
		name:         paramRefs.name,
		objectNumber: paramRefs.objectNumbers[0],
	}, nil
}

type parameterNumberArray struct {
	name   string
	values []float64
}

var regexArray = regexp.MustCompile(`(\d+\.*\d*)`)

func (param *parameter) parseNumberArray() (parameterNumberArray, error) {
	submatches := regexArray.FindAllSubmatch(param.value, -1)
	if submatches == nil {
		err := errInvalidFormat(param.name)

		return parameterNumberArray{}, err
	}

	values := make([]float64, 0, len(submatches))

	for _, submatch := range submatches {
		float, err := strconv.ParseFloat(string(submatch[1]), 64)
		if err != nil {
			return parameterNumberArray{}, err
		}

		values = append(values, float)
	}

	paramArray := parameterNumberArray{
		name:   param.name,
		values: values,
	}

	return paramArray, nil
}

type parameterString struct {
	name  string
	value string
}

// /Type /Page
type parameterName struct {
	name  string
	value string
}

type parameterBool struct {
	name  string
	value bool
}
