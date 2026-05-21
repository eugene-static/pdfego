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
	w       meter.MM
	h       meter.MM
	slots   []*Slot
	opts    Options
}

type Slot struct {
	core      *core.Core
	w         meter.MM
	h         meter.MM
	renderers []renderer
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
	t.core.StartDocument()

	buf := t.core.AddPage()
	page := t.core.Page()
	x0, y0 := page.X0Y0()
	x, y := x0, y0
	headerHeight := meter.MM(0)

	for _, block := range t.blocks {
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
			t.core.WritePage()
			t.core.AddPage()
			x = x0
			y = y0 + headerHeight
		}

		block.render(buf, x, y+block.opts.Indent)
		block = nil

		y += height
	}

	t.core.FinishDocument()
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

func Repeat[T any](tmpl *Template, items []T, blockFunc func(b *Block, item T)) {
	tmpl.blocks = slices.Grow(tmpl.blocks, len(items))
	for i := range items {
		b := tmpl.Block()
		blockFunc(b, items[i])
	}
}

func (b *Block) render(buf *buffer.Buffer, x, y meter.MM) {
	for i := range b.slots {
		b.slots[i].render(buf, x, y)

		x += b.slots[i].width() + b.opts.Spacing
	}
}

func (b *Block) height() meter.MM {
	if b.h == 0 {
		heights := make([]meter.MM, 0, len(b.slots))

		for i := range b.slots {
			heights = append(heights, b.slots[i].height())
		}

		b.h = slices.Max(heights) + b.opts.Indent
	}

	return b.h
}

func (b *Block) width() meter.MM {
	if b.w == 0 {
		for i := range b.slots {
			b.w += b.slots[i].width()
		}

		b.w += spacing(b.opts.Spacing, len(b.slots))
	}

	return b.w
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

func (s *Slot) Table(rowsNum int, columns []meter.MM) *Table {
	columnsNum := len(columns)

	table := Table{
		core:      s.core,
		columns:   columns,
		rows:      make([]Row, rowsNum),
		cellsPool: make([]cell, rowsNum*columnsNum),
		rowspans:  make([]uint8, columnsNum),
	}

	s.core.IncreaseCellsCount(rowsNum * columnsNum)

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
	if s.h == 0 {
		for i := range s.renderers {
			s.h += s.renderers[i].height()
		}
	}

	return s.h
}

func (s *Slot) width() meter.MM {
	if s.w == 0 {
		for i := range s.renderers {
			s.w += s.renderers[i].width()
		}
	}

	return s.w
}

func spacing(sp meter.MM, arrLen int) meter.MM {
	if arrLen < 2 {
		return meter.MM(0)
	}

	return sp * meter.MM(arrLen-1)
}
