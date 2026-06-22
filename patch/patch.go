package patch

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/eugene-static/pdfego"
	"github.com/eugene-static/pdfego/internal/buffer"
)

const (
	refTail                       = " 0 R"
	endobj                        = "endobj"
	regexParameterReferenceLayout = `/%s\s*\[*(\s*\d+\s\d\sR)+\s*(]|)`
)

var (
	regexStartXref = regexp.MustCompile(`startxref\s*(\d+)\s*%%EOF`)
	regexReference = regexp.MustCompile(`\s*(\d+)\s\d\sR+\s*`)
)

type Patcher struct {
	core      *pdfego.Core
	data      []byte
	buffer    *buffer.Buffer
	catalog   *object
	pages     *object
	pagesKids []*object
	xref      *xref
	trailer   *trailer
}

func New(core *pdfego.Core) *Patcher {
	return &Patcher{
		core: core,
	}
}

func (p *Patcher) Patch(data []byte) error {
	rs := bytes.NewReader(data)
	reader := bufio.NewReader(rs)

	err := validate(reader)
	if err != nil {
		return err
	}

	p.data = data

	err = p.parse()

	return nil
}

func (p *Patcher) parse() error {
	xrefIndex, err := p.parseTrailer()
	if err != nil {
		return err
	}

	err = p.parseXref(xrefIndex)
	if err != nil {
		return err
	}

	paramName := "Root"

	rootParam, err := parseParametersReference(p.trailer.body, paramName)
	if err != nil {
		return err
	}

	p.trailer.setParameter(paramName, rootParam)

	catalog, err := p.getObject("Catalog", rootParam.objectNumbers[0])
	if err != nil {
		return err
	}

	p.catalog = catalog

	paramName = "Pages"

	catalogPagesParam, err := parseParametersReference(catalog.body, paramName)
	if err != nil {
		return err
	}

	p.catalog.setParameter(paramName, catalogPagesParam)

	pages, err := p.getObject(paramName, catalogPagesParam.objectNumbers[0])
	if err != nil {
		return err
	}

	p.pages = pages

	paramName = "Kids"

	pagesKidsParam, err := parseParametersReference(pages.body, paramName)
	if err != nil {
		return err
	}

	pages.setParameter(paramName, pagesKidsParam)

	pagesKids := make([]*object, 0, len(pagesKidsParam.objectNumbers))

	for _, pageNumber := range pagesKidsParam.objectNumbers {
		page, err := p.getObject("Page", pageNumber)
		if err != nil {
			return err
		}

		pagesKids = append(pagesKids, page)
	}

	p.pagesKids = pagesKids

	return nil
}

func (p *Patcher) parseTrailer() (int, error) {
	trailerIndex := bytes.LastIndex(p.data, []byte("trailer"))
	if trailerIndex == -1 {
		err := errors.New("не удалось найти trailer")

		return 0, err
	}

	trailerBody := p.data[trailerIndex:]

	xrefValue := regexStartXref.FindSubmatch(trailerBody)
	if len(xrefValue) == 0 {
		err := errors.New("не удалось найти startxref")

		return 0, err
	}

	xrefIndex := bytesToInt(xrefValue[1])

	p.trailer = &trailer{
		index: trailerIndex,
		body:  trailerBody,
	}

	return xrefIndex, nil
}

func (p *Patcher) parseXref(index int) error {
	xrefBody := p.data[index : p.trailer.index+1]

	lines := bytes.Split(xrefBody, []byte("\n"))
	if len(lines) < 2 {
		err := errors.New("таблица xref пуста")

		return err
	}

	sum := bytes.Split(lines[1], []byte(" "))
	if len(sum) != 2 {
		err := errors.New("неверный формат суммы xref")

		return err
	}

	offset := bytesToInt(sum[0])
	count := bytesToInt(sum[1])

	if count != len(lines)-2 {
		err := errors.New("таблица xref повреждена")

		return err
	}

	objectsIndexes := make([]int, 0, count)

	for i := 2; i < len(lines); i++ {
		ref := bytes.Split(lines[i], []byte(" "))
		if len(ref) != 3 {
			err := errors.New("неверный формат таблицы xref")

			return err
		}

		objectIndex := bytesToInt(ref[0])

		objectsIndexes = append(objectsIndexes, objectIndex)
	}

	p.xref = &xref{
		index:          index,
		offset:         offset,
		count:          count,
		objectsIndexes: objectsIndexes,
	}

	return nil
}

func (p *Patcher) getObject(name string, number int) (*object, error) {
	startIndex := p.xref.getObjectIndex(number)
	if startIndex == -1 {
		err := fmt.Errorf("не удалось найти объект %s_%d", name, number)

		return nil, err
	}

	endIndex := bytes.Index(p.data[startIndex:], []byte(endobj))
	if endIndex == -1 {
		err := fmt.Errorf("не удалось найти конец объекта %s_%d", name, number)

		return nil, err
	}

	endIndex += startIndex

	obj := &object{
		name:   name,
		number: number,
		body:   p.data[startIndex : endIndex+len(endobj)],
	}

	return obj, nil
}

func parseParametersReference(body []byte, parameterName string) (parameterReference, error) {
	regex, err := regexp.Compile(fmt.Sprintf(regexParameterReferenceLayout, parameterName))
	if err != nil {
		return parameterReference{}, err
	}

	paramBodyCoordinates := regex.FindIndex(body)
	if len(paramBodyCoordinates) == 0 {
		err = fmt.Errorf("не удалось найти параметр %s", parameterName)

		return parameterReference{}, err
	}

	references := regexReference.FindAllSubmatch(body[paramBodyCoordinates[0]:paramBodyCoordinates[1]], -1)
	if len(references) == 0 {
		err = fmt.Errorf("неверный формат ссылок параметра %s", parameterName)

		return parameterReference{}, err
	}

	objectNumbers := make([]int, 0, len(references))

	for _, reference := range references {
		objectNumber := bytesToInt(reference[1])

		objectNumbers = append(objectNumbers, objectNumber)
	}

	param := parameterReference{
		name:          parameterName,
		objectNumbers: objectNumbers,
		start:         paramBodyCoordinates[0],
		end:           paramBodyCoordinates[1],
	}

	return param, nil
}

type object struct {
	name   string
	number int
	body   []byte
	params map[string]parameter
}

func (obj *object) setParameter(name string, param parameter) {
	if obj.params == nil {
		obj.params = make(map[string]parameter)
	}

	obj.params[name] = param
}

type trailer struct {
	index  int
	body   []byte
	params map[string]parameter
}

func (tr *trailer) setParameter(name string, param parameter) {
	if tr.params == nil {
		tr.params = make(map[string]parameter)
	}

	tr.params[name] = param
}

type xref struct {
	index          int
	offset         int
	count          int
	objectsIndexes []int
}

func (x *xref) getObjectIndex(objectNumber int) int {
	if objectNumber > len(x.objectsIndexes) {
		return -1
	}

	return x.objectsIndexes[objectNumber]
}

type parameter interface{}

type parameterReference struct {
	name          string
	objectNumbers []int
	start         int
	end           int
}

type parameterArray struct {
	name  string
	array []float64
	start int
	end   int
}

func validate(rs *bufio.Reader) error {
	pdf, err := rs.Peek(4)
	if err != nil {
		return err
	}

	if string(pdf) != "%PDF" {
		err := errors.New("неверный формат PDF")

		return err
	}

	return nil
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func objectHeader(number int) []byte {
	numberStr := strconv.Itoa(number)

	return []byte(numberStr + refTail)
}

func bytesToInt(b []byte) int {
	res := 0

	for _, x := range b {
		res = res*10 + int(x-'0')
	}

	return res
}
