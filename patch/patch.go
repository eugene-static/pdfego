package patch

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/eugene-static/pdfego"
	"github.com/eugene-static/pdfego/internal/buffer"
)

const (
	refTail = " 0 R"
)

type Patcher struct {
	core    *pdfego.Core
	buffer  *buffer.Buffer
	catalog *object
	pages   *object
	trailer *object
}

func New(core *pdfego.Core) *Patcher {
	return &Patcher{
		core: core,
	}
}

func (p *Patcher) Patch(data []byte) error {
	if string(data[1:4]) != "PDF" {
		err := errors.New("неверный формат PDF")

		return err
	}

	trailerIndex := bytes.LastIndex(data, []byte("trailer"))
	if trailerIndex == -1 {
		err := errors.New("не удалось найти trailer")

		return err
	}

	root, err := extractParametersReferenceObjectNumbers(data, trailerIndex, "/Root")
	if err != nil {
		return err
	}

	if len(root.objectNumbers) != 1 {
		err = fmt.Errorf("неверный формат ссылки %s", root.name)
	}

	catalogIndex := bytes.Index(data, objectHeader(root.objectNumbers[0]))
	if catalogIndex == -1 {
		err = errors.New("не удалось найти объект /Catalog")

		return err
	}

	pagesParam, err := extractParametersReferenceObjectNumbers(data, catalogIndex, "/Pages")
	if err != nil {
		return err
	}

	if len(pagesParam.objectNumbers) != 1 {
		err = fmt.Errorf("неверный формат ссылки %s", pagesParam.name)
	}

	catalog, err := newObject(root.objectNumbers[0], data, catalogIndex, pagesParam)
	if err != nil {
		return err
	}

	p.catalog = catalog

	pagesIndex := bytes.Index(data, objectHeader(pagesParam.objectNumbers[0]))
	if pagesIndex == -1 {
		err = errors.New("не удалось найти объект /Pages")

		return err
	}

	kidsParam, err := extractParametersReferenceObjectNumbers(data, pagesIndex, "/Kids")
	if err != nil {
		return err
	}

	pages, err := newObject(pagesParam.objectNumbers[0], data, pagesIndex, kidsParam)
	if err != nil {
		return err
	}

	p.pages = pages

	return nil
}

func findObject(data []byte, number int, paramName string) (*object, error) {
	index := bytes.Index(data, objectHeader(number))
	if index == -1 {
		err := fmt.Errorf("не удалось найти объект: %d", number)

		return nil, err
	}

	param, err := extractParametersReferenceObjectNumbers(data, index, paramName)
	if err != nil {
		return nil, err
	}

	obj, err := newObject(number, data, index, param)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func extractParametersReferenceObjectNumbers(body []byte, objectStartIndex int, parameter string) (parRef parameterReference, err error) {
	parRef = parameterReference{
		name:          parameter,
		objectNumbers: make([]int, 0),
	}

	parameterStartIndex := bytes.Index(body[objectStartIndex:], []byte(parameter))
	if parameterStartIndex == -1 {
		err := fmt.Errorf("не удалось найти параметр %s", parameter)

		return parameterReference{}, err
	}

	parRef.startIndex = parameterStartIndex

	parameterEndByte := byte(']')

	startIndex := bytes.IndexByte(body[parameterStartIndex:], '[')
	if startIndex == -1 {
		startIndex += parameterStartIndex + len(parameter)

		for !isDigit(body[startIndex]) {
			startIndex++
		}

		endIndex := startIndex

		for i := startIndex; isDigit(body[i]) && i < len(body); i++ {
			endIndex++
		}

		objectNumber, err := strconv.Atoi(string(body[startIndex:endIndex]))
		if err != nil {
			return parameterReference{}, err // ошибки быть не должно по логике
		}

		parameterEndByte = 'R'

		for endIndex < len(body) && body[endIndex] != parameterEndByte {
			endIndex++
		}

		parRef.endIndex = endIndex + 1

		parRef.objectNumbers = append(parRef.objectNumbers, objectNumber)

		return parRef, nil
	}

	i := startIndex + parameterStartIndex

	for i < len(body) && body[i] != parameterEndByte {
		if isDigit(body[i]) {
			j := i + 1

			for j < len(body) && body[j] != ' ' {
				j++
			}

			objectNumber, err := strconv.Atoi(string(body[i:j]))
			if err != nil {
				return parameterReference{}, err
			}

			parRef.objectNumbers = append(parRef.objectNumbers, objectNumber)

			i = j + len(refTail)

			continue
		}

		i++

	}

	parRef.endIndex = i + 1

	return parRef, nil
}

type object struct {
	number int
	body   []byte
	param  parameterReference
}

type parameterReference struct {
	name          string
	objectNumbers []int
	startIndex    int
	endIndex      int
}

func newObject(number int, body []byte, startIndex int, param parameterReference) (*object, error) {
	endIndex := bytes.Index(body[startIndex:], []byte("endobj"))
	if endIndex == -1 {
		err := fmt.Errorf("не удалось найти конец объекта %d", number)

		return nil, err
	}

	endIndex += startIndex

	obj := &object{
		number: number,
		body:   body[startIndex:endIndex],
		param:  param,
	}

	return obj, nil
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func objectHeader(number int) []byte {
	numberStr := strconv.Itoa(number)

	return []byte(numberStr + refTail)
}
