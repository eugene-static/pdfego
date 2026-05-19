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
	profile byte
	slots   []*Slot
	opts    Options
}

type Slot struct {
	core      *core.Core
	renderers []renderer
}

type Table struct {
	core    *core.Core
	columns []meter.MM
	rows    []*Row
	opts    Options
}

type Options struct {
	Spacing meter.MM
	Indent  meter.MM
	Ledge   meter.MM
}

type renderer interface {
	render(b *buffer.Buffer, x, y meter.MM)
	height() meter.MM
	width() meter.MM
	options() Options
}

func New(core *core.Core) *Template {
	return &Template{
		core: core,
	}
}

func (t *Template) Block(options ...Options) *Block {
	var opts Options
	if len(options) > 0 {
		opts = options[0]
	}

	block := t.block(defaultBlock, opts)

	return block
}

func (t *Template) Header(options ...Options) *Block {
	var opts Options
	if len(options) > 0 {
		opts = options[0]
	}

	block := t.block(headerBlock, opts)

	return block
}

func (t *Template) EndHeader(options ...Options) *Block {
	var opts Options
	if len(options) > 0 {
		opts = options[0]
	}

	block := t.block(headerStopBlock, opts)

	return block
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

		y += block.opts.Indent

		block.render(buf, x, y)

		y += height
	}

	t.core.FillBuffer()
}

func (t *Template) Bytes() []byte {
	return t.core.Bytes()
}

func (t *Template) block(profile byte, options Options) *Block {
	block := Block{
		core:    t.core,
		profile: profile,
		opts:    options,
	}

	t.blocks = append(t.blocks, &block)

	return &block
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

		x += b.slots[i].width() + b.opts.Spacing
	}
}

func (b *Block) height() meter.MM {
	heights := make([]meter.MM, 0, len(b.slots))

	for i := range b.slots {
		heights = append(heights, b.slots[i].height())
	}

	return slices.Max(heights) + b.opts.Indent
}

func (b *Block) width() meter.MM {
	width := spacing(b.opts.Spacing, len(b.slots))

	for i := range b.slots {
		width += b.slots[i].width()
	}

	return width
}

func (b *Block) options() Options {
	return b.opts
}

func (s *Slot) Block(options ...Options) *Block {
	var opts Options
	if len(options) > 0 {
		opts = options[0]
	}

	block := Block{
		core: s.core,
		opts: opts,
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
		opts := s.renderers[i].options()

		y += opts.Indent

		s.renderers[i].render(buf, x, y)

		y += s.renderers[i].height()
	}
}

func (s *Slot) height() meter.MM {
	height := meter.MM(0)

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

		for i, c := range row.cells {
			c.x = cellX
			c.y = row.y
			cellX += t.columns[i]

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

	return h + t.options().Indent
}

func (t *Table) width() (w meter.MM) {
	for i := range t.columns {
		w += t.columns[i]
	}

	return w
}

func (t *Table) options() Options {
	return t.opts
}

func (t *Table) cellHeight(rowIndex, rowspan int) (h meter.MM) {
	for i := rowIndex; i < min(len(t.rows), rowIndex+rowspan); i++ {
		h += t.rows[i].height
	}

	return h
}

func spacing(sp meter.MM, arrLen int) meter.MM {
	if arrLen < 2 {
		return meter.MM(0)
	}

	return sp * meter.MM(arrLen-1)
}
