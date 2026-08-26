package pages

import (
	"github.com/eugene-static/pdfego/internal/components/object"
	"github.com/eugene-static/pdfego/internal/components/parameter"
	"github.com/eugene-static/pdfego/internal/components/stream"
	"github.com/eugene-static/pdfego/internal/compressor"
	"github.com/eugene-static/pdfego/unit"
)

const (
	ObjectNumber parameter.Reference = 1

	A4Width  unit.PT = 595.2
	A4Height unit.PT = 841.89

	AliasPagesCount = "PagesCount"
)

type Pages struct {
	compressor *compressor.Compressor
	config     Config
	contents   contents
	objects    objects
}

func New(comp *compressor.Compressor, config *Config) *Pages {
	return &Pages{
		compressor: comp,
		contents: contents{
			references: make([]parameter.Reference, 0, 3),
			page:       object.New(),
			watermark:  object.New(),
			header:     object.New(),
		},
		objects: objects{
			parent:     object.NewSimple(),
			pagesCount: object.New(),
			kids:       make([]*object.Object, 0, 1),
		},
		config: *config,
	}
}

func (p *Pages) X0() unit.MM {
	return p.config.LowerLeftX + p.config.MarginLeft
}

func (p *Pages) Y0() unit.MM {
	return p.config.MarginTop - p.config.UpperRightY
}

func (p *Pages) Config() Config {
	return p.config
}

func (p *Pages) NewPage() {
	obj := object.NewSimple()

	p.objects.kids = append(p.objects.kids, obj)

	p.contents.references = p.contents.references[:0]

	if p.contents.header.Number() > 0 {
		p.contents.references = append(p.contents.references, p.contents.header.Number())
	}

	p.contents.page.Stream().Reset()
}

func (p *Pages) PageStream() *stream.Stream {
	return p.contents.page.Stream()
}

func (p *Pages) WatermarkStream() *stream.Stream {
	p.contents.watermark.Stream().Reset()

	return p.contents.watermark.Stream()
}

func (p *Pages) WatermarkRelease() {
	p.contents.watermark.Reset()
}

func (p *Pages) HeaderStream() *stream.Stream {
	p.contents.header.Stream().Reset()

	return p.contents.header.Stream()
}

func (p *Pages) HeaderRelease() {
	p.contents.header.Reset()
}

func (p *Pages) CountStream() *stream.Stream {
	p.objects.pagesCount.Stream().Reset()

	return p.objects.pagesCount.Stream()
}

func (p *Pages) Count() int {
	return len(p.objects.kids)
}

type Config struct {
	LowerLeftX   unit.MM
	LowerLeftY   unit.MM
	UpperRightX  unit.MM
	UpperRightY  unit.MM
	MarginLeft   unit.MM
	MarginTop    unit.MM
	MarginRight  unit.MM
	MarginBottom unit.MM
}

func NewConfig(width, height unit.MM) *Config {
	return &Config{
		LowerLeftX:   0,
		LowerLeftY:   0,
		UpperRightX:  width,
		UpperRightY:  height,
		MarginLeft:   3,
		MarginTop:    3,
		MarginRight:  3,
		MarginBottom: 8,
	}
}

func (cfg *Config) SetMargins(left, top, right, bottom unit.MM) {
	cfg.MarginLeft = left
	cfg.MarginTop = top
	cfg.MarginRight = right
	cfg.MarginBottom = bottom
}

func (cfg *Config) bBox() parameter.Numbers {
	return parameter.Numbers{
		parameter.Number(cfg.LowerLeftX.PT()),
		parameter.Number(cfg.LowerLeftY.PT()),
		parameter.Number(cfg.UpperRightX.PT()),
		parameter.Number(cfg.UpperRightY.PT()),
	}
}

type ViewBox struct {
	X0     unit.MM
	Y0     unit.MM
	Width  unit.MM
	Height unit.MM
}

func (p *Pages) ViewBox() ViewBox {
	return ViewBox{
		X0:     p.X0(),
		Y0:     p.Y0(),
		Width:  p.config.UpperRightX - p.config.MarginRight - p.config.MarginLeft,
		Height: p.config.UpperRightY - p.config.MarginTop - p.config.MarginBottom,
	}
}
