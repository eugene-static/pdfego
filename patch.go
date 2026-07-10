package pdfego

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"regexp"

	"github.com/eugene-static/pdfego/internal/buffer"
)

const (
	refTail = " 0 R"
	endobj  = "endobj"
)

var (
	regexStartXref = regexp.MustCompile(`startxref\s*(\d+)\s*%%EOF`)
)

type Patcher struct {
	core      *Core
	data      []byte
	buffer    *buffer.Buffer
	catalog   *object
	pages     *object
	pagesKids []*object
	resources *object
	xref      *xref
	trailer   *trailer
}

func NewPatcher(core *Core) *Patcher {
	return &Patcher{
		core: core,
	}
}

func (p *Patcher) Watermark(options ...WatermarkOptions) *Watermark {
	opts := getOptions(options)
	alignH, alignV := parseAlignment(opts.Align)
	x0, y0 := p.core.page.x0y0()

	return &Watermark{
		block: &Block{
			core: p.core,
		},
		buf:    p.core.page.newWatermark(),
		x0:     x0,
		y0:     y0,
		alignH: alignH,
		alignV: alignV,
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
	if err != nil {
		return err
	}

	return nil
}

func (p *Patcher) parse() error {
	_trailer, err := parseTrailer(p.data)
	if err != nil {
		return err
	}

	_xref, err := parseXref(p.data, _trailer.index)
	if err != nil {
		return err
	}

	rootParam, ok := _trailer.dictionary.parameter("Root")
	if !ok {
		err = errNotFound("Root")

		return err
	}

	rootReference, err := rootParam.parseReference()
	if err != nil {
		return err
	}

	_catalog, err := p.getObject("Catalog", rootReference.objectNumber)
	if err != nil {
		return err
	}

	pagesParam, ok := _catalog.dictionary.parameter("Pages")
	if !ok {
		err = errNotFound("Pages")

		return err
	}

	pagesReference, err := pagesParam.parseReference()
	if err != nil {
		return err
	}

	_pages, err := p.getObject("Pages", pagesReference.objectNumber)
	if err != nil {
		return err
	}

	pagesKidsParam, ok := _pages.dictionary.parameter("Kids")
	if !ok {
		err = errNotFound("Kids")

		return err
	}

	pagesKidsReferences, err := pagesKidsParam.parseReferences()
	if err != nil {
		return err
	}

	pagesKids := make([]*object, 0, len(pagesKidsReferences.objectNumbers))

	for _, pageNumber := range pagesKidsReferences.objectNumbers {
		_page, err := p.getObject("Page", pageNumber)
		if err != nil {
			return err
		}

		pagesKids = append(pagesKids, _page)
	}

	p.trailer = _trailer
	p.xref = _xref
	p.catalog = _catalog
	p.pages = _pages
	p.pagesKids = pagesKids

	return nil
}

func (p *Patcher) getResources() (object, error) {
	resourcesParameters := make(map[string][]byte)

	for _, obj := range p.pagesKids {
		resourcesParam, ok := obj.dictionary.parameter("Resources")
		if ok {
			resources, err := p.resolve(resourcesParam)
			if err != nil {
				return object{}, err
			}

		}
	}

	return p.resources, nil
}

func (p *Patcher) resolveUp(dict dictionary, name string) (parameter, error) {
	param, ok := dict.parameter(name)
	if ok {
		return param, nil
	}

	parent, ok := dict.parameter("Parent")
	if !ok {
		err := errNotFound("Parent reference")

		return parameter{}, err
	}

	dict, err := p.resolve(parent)
	if err != nil {
		return parameter{}, err
	}

	param, err = p.resolveUp(dict, name)
	if err != nil {
		return parameter{}, err
	}

	return param, nil
}

func (p *Patcher) resolve(parameter parameter) (dictionary, error) {
	params := parseDictionary(parameter.value)
	if len(params) != 0 {
		return params, nil
	}

	paramsReference, err := parameter.parseReference()
	if err != nil {
		return params, err
	}

	obj, err := p.getObject("", paramsReference.objectNumber)
	if err != nil {
		return params, err
	}

	return obj.dictionary, nil
}

func (p *Patcher) prepare() error {
	p.core.offsets = p.xref.offsets

	_, err := p.core.mainBuffer.Write(p.data)
	if err != nil {
		return err
	}

	return nil
}

func (p *Patcher) patch() error {
	return nil
}

func (p *Patcher) getObject(name string, number int) (*object, error) {
	startIndex := p.xref.getOffset(number)
	if startIndex == -1 {
		err := errNotFound(fmt.Sprintf("%s_%d", name, number))

		return nil, err
	}

	endIndex := bytes.Index(p.data[startIndex:], []byte(endobj))
	if endIndex == -1 {
		err := fmt.Errorf("не удалось найти конец объекта %s_%d", name, number)

		return nil, err
	}

	endIndex += startIndex

	objBody := p.data[startIndex:endIndex]

	startStream, endStream := parseStream(objBody)
	stream := objBody[startStream:endStream]

	dict := parseDictionary(objBody[startIndex:startStream])

	obj := &object{
		name:       name,
		number:     number,
		stream:     stream,
		dictionary: dict,
	}

	return obj, nil
}

type object struct {
	name       string
	number     int
	stream     []byte
	dictionary dictionary
}

type notFoundError struct {
	parameter string
}

func (e notFoundError) Error() string {
	return "не удалось найти " + e.parameter
}

func errNotFound(parameter string) error {
	err := notFoundError{parameter}

	return err
}

type invalidFormatError struct {
	parameter string
}

func (e invalidFormatError) Error() string {
	return "неверный формат параметра " + e.parameter
}

func errInvalidFormat(parameter string) error {
	err := invalidFormatError{parameter}

	return err
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

func bytesToInt(b []byte) int {
	res := 0

	for _, x := range b {
		res = res*10 + int(x-'0')
	}

	return res
}
