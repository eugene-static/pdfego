package core

import (
	"github.com/eugene-static/pdf-craft/buffer"
	"github.com/eugene-static/pdf-craft/meter"
)

type Page struct {
	x      meter.MM
	y      meter.MM
	width  meter.MM
	height meter.MM
	margin meter.MM
}

func (p *Page) X0Y0() (meter.MM, meter.MM) {
	return p.margin, p.height - p.margin
}

func (p *Page) XY() (meter.MM, meter.MM) {
	return p.x, p.y
}

func (p *Page) IsBelowBottomBorder(val meter.MM) bool {
	return (val + p.margin) > 0
}

func (core *Core) SetMargin(margin float64) {
	core.page.margin = meter.MM(margin)
}

func (core *Core) AddPage() (*buffer.Buffer, Page) {
	buf := buffer.New()

	if core.headBuffer != nil && core.headBuffer.Len() > 0 {
		buf.Write(core.headBuffer.Bytes())
	}

	core.pageBuffers = append(core.pageBuffers, buf)

	return buf, core.page
}

func (core *Core) AddHeader() *buffer.Buffer {
	if core.headBuffer != nil {
		core.headBuffer.Reset()

		return core.headBuffer
	}

	buf := buffer.New()

	core.headBuffer = buf

	return buf
}

func (core *Core) RemoveHeader() {
	core.headBuffer.Reset()
}

func (core *Core) pageBottomEdge() meter.MM {
	return core.page.height - core.page.margin
}

func (core *Core) pageRightEdge() meter.MM {
	return core.page.width - core.page.margin
}
