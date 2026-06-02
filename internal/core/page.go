package core

import (
	"github.com/eugene-static/pdf-craft/internal/buffer"
	"github.com/eugene-static/pdf-craft/pkg/unit"
)

type Page struct {
	buffer       *buffer.Buffer
	headerBuffer *buffer.Buffer
	footerBuffer *buffer.Buffer
	width        unit.MM
	height       unit.MM
	margin       unit.MM
	marginLeft   unit.MM
	marginRight  unit.MM
	marginTop    unit.MM
	marginBottom unit.MM //TODO:
}

func (p *Page) X0Y0() (unit.MM, unit.MM) {
	return p.margin, p.margin - p.height
}

func (p *Page) IsBelowBottomBorder(y unit.MM) bool {
	return y+p.margin > 0
}

func (p *Page) New() error {
	if p.headerBuffer != nil && p.headerBuffer.Len() > 0 {
		_, err := p.buffer.Write(p.headerBuffer.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Page) Buffer() *buffer.Buffer {
	return p.buffer
}

func (p *Page) AddHeader() *buffer.Buffer {
	if p.headerBuffer != nil {
		p.headerBuffer.Reset()

		return p.headerBuffer
	}

	buf := buffer.New(bufferSize)

	p.headerBuffer = buf

	return buf
}

func (p *Page) RemoveHeader() {
	p.headerBuffer.Reset()
}

func (core *Core) RenderPage() {
	core.writePage()
}
