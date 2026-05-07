package template

type page struct {
	width, height float64
	margin        float64
}

func (core *Core) SetMargin(margin float64) {
	core.page.margin = margin
}

func (core *Core) addPage() *buffer {
	core.setXY(core.x0y0())

	buf := newBuffer()

	if core.headBuffer != nil && core.headBuffer.content.Len() > 0 {
		buf.content.Write(core.headBuffer.content.Bytes())
	}

	core.pageBuffers = append(core.pageBuffers, buf)

	return buf
}
