package pages

import (
	"github.com/eugene-static/pdfego/internal/components/dictionary"
	"github.com/eugene-static/pdfego/internal/components/object"
	"github.com/eugene-static/pdfego/internal/components/object/resources"
	"github.com/eugene-static/pdfego/internal/components/parameter"
)

type contents struct {
	references []parameter.Reference
	watermark  *object.Object
	header     *object.Object
	page       *object.Object
}

type objects struct {
	parent     *object.Object
	pagesCount *object.Object
	kids       []*object.Object
}

type pages struct {
	Type     parameter.Name       `pdf:"Type"`
	Kids     parameter.References `pdf:"Kids"`
	Count    parameter.Integer    `pdf:"Count"`
	MediaBox parameter.Numbers    `pdf:"MediaBox"`
}

type page struct {
	Type      parameter.Name       `pdf:"Type"`
	Parent    parameter.Reference  `pdf:"Parent"`
	Resources parameter.Reference  `pdf:"Resources"`
	Contents  parameter.References `pdf:"Contents"`
}

type pagesCount struct {
	Type      parameter.Name      `pdf:"Type"`
	Subtype   parameter.Name      `pdf:"Subtype"`
	FormType  parameter.Integer   `pdf:"FormType"`
	Resources parameter.Reference `pdf:"Resources"`
	BBox      parameter.Numbers   `pdf:"BBox"`
	Filter    parameter.Name      `pdf:"Filter,omitempty"`
	Length    parameter.Integer   `pdf:"Length"`
}

func (p *Pages) Pages() *object.Object {
	kidsCount := len(p.objects.kids)

	kids := make(parameter.References, 0, kidsCount)

	for _, kid := range p.objects.kids {
		kids = append(kids, kid.Number())
	}

	_dictionary := &pages{
		Type:     parameter.Name("Pages"),
		Kids:     kids,
		Count:    parameter.Integer(kidsCount),
		MediaBox: p.config.bBox(),
	}

	p.objects.parent.ReadDictionary(_dictionary)
	p.objects.parent.SetNumber(ObjectNumber)

	return p.objects.parent
}

func (p *Pages) Page() *object.Object {
	obj := p.objects.kids[len(p.objects.kids)-1]

	if p.contents.watermark.Number() > 0 {
		p.contents.references = append(p.contents.references, p.contents.watermark.Number())
	}

	p.contents.references = append(p.contents.references, p.contents.page.Number())

	_dictionary := page{
		Type:      "Page",
		Parent:    ObjectNumber,
		Resources: resources.ObjectNumber,
		Contents:  p.contents.references,
	}

	obj.ReadDictionary(_dictionary)

	return obj
}

func (p *Pages) PagesCount(cfg parameter.Numbers) (*object.Object, error) {
	var filter parameter.Name

	if p.compressor.Enabled() {
		filter = parameter.FlateDecode
	}

	bytes, err := p.compressor.Compress(p.objects.pagesCount.Stream().Bytes())
	if err != nil {
		return nil, err
	}

	p.objects.pagesCount.Stream().Reset()

	_dictionary := pagesCount{
		Type:      "XObject",
		Subtype:   "Form",
		FormType:  1,
		Resources: resources.ObjectNumber,
		BBox:      cfg,
		Filter:    filter,
		Length:    parameter.Integer(len(bytes)),
	}

	p.objects.pagesCount.ReadDictionary(_dictionary)
	p.objects.pagesCount.ReadStream(bytes)

	return p.objects.pagesCount, nil
}

func (p *Pages) HeaderContent() (*object.Object, error) {
	if p.contents.header.Stream().Len() > 0 {
		err := p.content(p.contents.header)
		if err != nil {
			return nil, err
		}
	}

	return p.contents.header, nil
}

func (p *Pages) WatermarkContent() (*object.Object, error) {
	if p.contents.watermark.Stream().Len() > 0 {
		err := p.content(p.contents.watermark)
		if err != nil {
			return nil, err
		}
	}

	return p.contents.watermark, nil
}

func (p *Pages) PageContent() (*object.Object, error) {
	if p.contents.page.Stream().Len() > 0 {
		err := p.content(p.contents.page)
		if err != nil {
			return nil, err
		}
	}

	return p.contents.page, nil
}

func (p *Pages) content(obj *object.Object) error {
	var filter parameter.Name

	if p.compressor.Enabled() {
		filter = parameter.FlateDecode
	}

	bytes, err := p.compressor.Compress(obj.Stream().Bytes())
	if err != nil {
		return err
	}

	obj.Stream().Reset()

	_dictionary := dictionary.Content{
		Filter: filter,
		Length: parameter.Integer(len(bytes)),
	}

	obj.ReadDictionary(_dictionary)
	obj.ReadStream(bytes)

	return nil
}
