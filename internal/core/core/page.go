package core

import (
	"github.com/eugene-static/pdf-craft/internal/buffer"
	"github.com/eugene-static/pdf-craft/pkg/meter"
)

type Page struct {
	width  meter.MM
	height meter.MM
	margin meter.MM
}

func (p Page) X0Y0() (meter.MM, meter.MM) {
	return p.margin, p.margin - p.height
}

func (p Page) IsBelowBottomBorder(y meter.MM) bool {
	return y+p.margin > 0
}

func (core *Core) Page() Page {
	return core.page
}

func (core *Core) SetMargin(margin float64) {
	core.page.margin = meter.MM(margin)
}

func (core *Core) NewPage() *buffer.Buffer {
	if core.pageBuffer.Len() > 0 {
		core.writePage()
		core.pageBuffer.Reset()
	}

	if core.headBuffer != nil && core.headBuffer.Len() > 0 {
		core.pageBuffer.Write(core.headBuffer.Bytes())
	}

	return core.pageBuffer
}

func (core *Core) AddHeader() *buffer.Buffer {
	if core.headBuffer != nil {
		core.headBuffer.Reset()

		return core.headBuffer
	}

	buf := buffer.New(64 * 1024)

	core.headBuffer = buf

	return buf
}

func (core *Core) RemoveHeader() {
	core.headBuffer.Reset()
}
