package template

import (
	"slices"

	"github.com/eugene-static/pdf-craft/buffer"
	core "github.com/eugene-static/pdf-craft/core"
	"github.com/eugene-static/pdf-craft/meter"
)

const (
	defaultBlock = iota
	headerBlock
	headerStopBlock
)

type Template struct {
	core   *core.Core
	blocks []*Block
}

type Block struct {
	core    *core.Core
	slots   []*Slot
	profile byte
}

type Slot struct {
	core      *core.Core
	renderers []renderer
}

type Table struct {
	core    *core.Core
	columns []meter.MM
	rows    []*Row
}

type renderer interface {
	render(b *buffer.Buffer, x, y meter.MM)
	height() meter.MM
	width() meter.MM
}

func New(core *core.Core) *Template {
	return &Template{
		core: core,
	}
}

func (t *Template) Block() *Block {
	block := t.block(defaultBlock)

	return block
}

func (t *Template) Header() *Block {
	block := t.block(headerBlock)

	return block
}

func (t *Template) EndHeader() *Block {
	block := t.block(headerStopBlock)

	return block
}

func (t *Template) block(profile byte) *Block {
	block := Block{
		core:    t.core,
		profile: profile,
	}

	t.blocks = append(t.blocks, &block)

	return &block
}

func (t *Template) Render() {
	buf := t.core.AddPage()
	page := t.core.Page()
	x0, y0 := page.X0Y0()
	x, y := x0, y0

	for _, block := range t.blocks {
		headerHeight := meter.MM(0)
		height := block.height()

		switch block.profile {
		case headerBlock:
			headerBuf := t.core.AddHeader()

			block.render(headerBuf, x0, y0)

			headerHeight = height
		case headerStopBlock:
			t.core.RemoveHeader()

			headerHeight = 0
		default:
			//
		}

		if (page.Margin() + y + height) > 0 {
			buf = t.core.AddPage()
			x = x0
			y = y0 + headerHeight
		}

		block.render(buf, x, y)

		y += height
	}

	t.core.FillBuffer()
}

func (t *Template) Bytes() []byte {
	return t.core.Bytes()
}

func (b *Block) Slot() *Slot {
	slot := Slot{
		core: b.core,
	}

	b.slots = append(b.slots, &slot)

	return &slot
}

func (b *Block) Add(blockFunc func(*Block)) *Block {
	blockFunc(b)

	return b
}

func (b *Block) render(buf *buffer.Buffer, x, y meter.MM) {
	for i := range b.slots {
		b.slots[i].render(buf, x, y)

		x += b.slots[i].width()
	}
}

func (b *Block) height() meter.MM {
	heights := make([]meter.MM, 0, len(b.slots))

	for i := range b.slots {
		heights = append(heights, b.slots[i].height())
	}

	return slices.Max(heights)
}

func (b *Block) width() meter.MM {
	width := meter.MM(0.0)

	for i := range b.slots {
		width += b.slots[i].width()
	}

	return width
}

func (s *Slot) Block() *Block {
	block := Block{
		core: s.core,
	}

	s.renderers = append(s.renderers, &block)

	return &block
}

func (s *Slot) Table(column meter.MM, columns ...meter.MM) *Table {
	cols := make([]meter.MM, len(columns)+1)
	cols[0] = column
	copy(cols[1:], columns)

	table := Table{
		core:    s.core,
		columns: cols,
	}

	s.renderers = append(s.renderers, &table)

	return &table
}

func (s *Slot) Add(slotFunc func(s *Slot)) *Slot {
	slotFunc(s)

	return s
}

func (s *Slot) render(buf *buffer.Buffer, x, y meter.MM) {
	for i := range s.renderers {
		s.renderers[i].render(buf, x, y)

		y += s.renderers[i].height()
	}
}

func (s *Slot) height() meter.MM {
	height := meter.MM(0.0)

	for i := range s.renderers {
		height += s.renderers[i].height()
	}

	return height
}

func (s *Slot) width() meter.MM {
	width := meter.MM(0.0)

	for i := range s.renderers {
		width += s.renderers[i].width()
	}

	return width
}

func (t *Table) Row() *Row {
	row := t.newRow()

	t.rows = append(t.rows, row)

	return row
}

func (t *Table) Add(rowFunc func(t *Table)) {
	rowFunc(t)
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
		height:     meter.FontHeight(t.core.DefaultFontSize()),
		columns:    t.columns,
		columnsLen: columnsLen,
		cells:      cells,
		rowspans:   rowspans,
	}

	return r
}

func (t *Table) render(buf *buffer.Buffer, x, y meter.MM) {
	height := meter.MM(0.0)

	for rowIndex, row := range t.rows {
		row.x = x
		row.y = y + height
		height += row.height

		cellX := row.x

		for _, c := range row.cells {
			c.x = cellX
			c.y = row.y
			cellX += c.width

			if !c.busy {
				continue
			}

			c.height = t.cellHeight(rowIndex, c.rowspan)

			c.render(buf)
		}
	}
}

func (t *Table) height() (h meter.MM) {
	for i := range t.rows {
		h += t.rows[i].height
	}

	return h
}

func (t *Table) width() (w meter.MM) {
	for i := range t.columns {
		w += t.columns[i]
	}

	return w
}

func (t *Table) cellHeight(rowIndex, rowspan int) (h meter.MM) {
	for i := rowIndex; i < min(len(t.rows), rowIndex+rowspan); i++ {
		h += t.rows[i].height
	}

	return h
}
