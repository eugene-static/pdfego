package template

import (
	"slices"
)

const (
	defaultFrame = iota
	headerFrame
	headerStopFrame
)

type Template struct {
	core   *Core
	frames []Frame
}

type Frame struct {
	core    *Core
	fields  []Field
	profile byte
}

type Field struct {
	core      *Core
	renderers []renderer
}

type Table struct {
	core    *Core
	columns []float64
	rows    []*Row
}

type renderer interface {
	render(b *buffer, x, y float64)
	height() float64
	width() float64
}

func (t *Template) Frame() *Frame {
	frame := t.frame(defaultFrame)

	return frame
}

func (t *Template) Header() *Frame {
	frame := t.frame(headerFrame)

	return frame
}

func (t *Template) EndHeader() *Frame {
	frame := t.frame(headerStopFrame)

	return frame
}

func (t *Template) frame(profile byte) *Frame {
	frame := Frame{
		profile: profile,
	}

	t.frames = append(t.frames, frame)

	return &frame
}

func (t *Template) Render() {
	buf := t.core.addPage()

	for i := range t.frames {
		headerHeight := 0.0

		if t.frames[i].profile == headerFrame {
			x, y := t.core.x0y0()

			t.frames[i].render(t.core.headBuffer, x, y)

			headerHeight = t.frames[i].height()
		}

		if t.frames[i].profile == headerStopFrame {
			t.core.headBuffer.reset()

			headerHeight = 0
		}

		x, y := t.core.xy()
		height := t.frames[i].height()

		if y+height > t.core.page.height {
			buf = t.core.addPage()

			x, y = t.core.xy()
			y += headerHeight
		}

		t.frames[i].render(buf, x, y)

		t.core.setXY(x, y+height)
	}

	t.core.writePages()
}

func (f *Frame) Field() *Field {
	field := Field{
		core: f.core,
	}

	f.fields = append(f.fields, field)

	return &field
}

func (f *Frame) render(buf *buffer, x, y float64) {
	for i := range f.fields {
		f.fields[i].render(buf, x, y)

		x += f.fields[i].width()
	}
}

func (f *Frame) height() float64 {
	heights := make([]float64, 0, len(f.fields))

	for i := range f.fields {
		heights = append(heights, f.fields[i].height())
	}

	return slices.Max(heights)
}

func (f *Frame) width() float64 {
	width := 0.0

	for i := range f.fields {
		width += f.fields[i].width()
	}

	return width
}

func (f *Field) Frame() *Frame {
	frame := Frame{
		core: f.core,
	}

	f.renderers = append(f.renderers, &frame)

	return &frame
}

func (f *Field) Table(column float64, columns ...float64) *Table {
	cols := make([]float64, len(columns)+1)
	cols[0] = column
	copy(cols[1:], columns)

	table := Table{
		core:    f.core,
		columns: cols,
	}

	f.renderers = append(f.renderers, &table)

	return &table
}

func (f *Field) render(buf *buffer, x, y float64) {
	for i := range f.renderers {
		f.renderers[i].render(buf, x, y)

		y += f.renderers[i].height()
	}
}

func (f *Field) height() float64 {
	height := 0.0

	for i := range f.renderers {
		height += f.renderers[i].height()
	}

	return height
}

func (f *Field) width() float64 {
	width := 0.0

	for i := range f.renderers {
		width += f.renderers[i].width()
	}

	return width
}

func (t *Table) Row() *Row {
	row := t.newRow()

	t.rows = append(t.rows, row)

	return row
}

func (t *Table) newRow() *Row {
	columnsLen := len(t.columns)
	cells := make([]cell, columnsLen)
	rowspans := make([]int, columnsLen)

	rowsLen := len(t.rows)

	if rowsLen > 0 {
		copy(rowspans, t.rows[rowsLen-1].decrementRowSpans())
	}

	r := &Row{
		core:       t.core,
		height:     t.core.fontHeight,
		columns:    t.columns,
		columnsLen: columnsLen,
		cells:      cells,
		rowspans:   rowspans,
	}

	return r
}

func (t *Table) render(buf *buffer, x, y float64) {
	height := 0.0

	for rowIndex, row := range t.rows {
		row.x = x
		row.y = y + height

		cellX := row.x

		for _, c := range row.cells {
			c.x = cellX
			c.y = row.y
			cellX += c.width

			if !c.busy {
				continue
			}

			c.height = t.cellHeight(rowIndex, c.rowspan)
			height += row.height

			c.render(buf)
		}
	}
}

func (t *Table) height() (h float64) {
	for i := range t.rows {
		h += t.rows[i].height
	}

	return h
}

func (t *Table) width() (w float64) {
	for i := range t.columns {
		w += t.columns[i]
	}

	return w
}

func (t *Table) cellHeight(rowIndex, rowspan int) (h float64) {
	for i := rowIndex; i < min(len(t.rows), rowIndex+rowspan); i++ {
		h += t.rows[i].height
	}

	return h
}
