package template

import (
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
	core         *core.Core
	buf          *buffer.Buffer
	page         core.Page
	block        *Block
	headerHeight meter.MM
	x0           meter.MM
	y0           meter.MM
	x            meter.MM
	y            meter.MM
}

type Block struct {
	profile byte
	core    *core.Core
	buf     *buffer.Buffer
	x0      meter.MM
	y0      meter.MM
	x       meter.MM
	y       meter.MM
	w       meter.MM
	h       meter.MM
	slots   []*Slot
	slotV2  *Slot
	opts    Options
}

type Slot struct {
	core    *core.Core
	buf     *buffer.Buffer
	x       meter.MM
	y       meter.MM
	w       meter.MM
	h       meter.MM
	printer printer
	block   *Block
	table   *Table

	renderers []printer
}

type Options struct {
	Spacing meter.MM
	Indent  meter.MM
	Ledge   meter.MM
}

type printer interface {
	print(x, y meter.MM)
	height() meter.MM
	width() meter.MM
	options() Options
}

func New(core *core.Core) *Template {
	core.StartDocument()
	page := core.Page()
	x0, y0 := page.X0Y0()

	return &Template{
		core: core,
		buf:  core.AddPage(),
		page: page,
		x0:   x0,
		y0:   y0,
		x:    x0,
		y:    y0,
	}
}

func (t *Template) Block(options ...Options) *Block {
	var opts Options
	if len(options) > 0 {
		opts = options[0]
	}

	block := t.newBlock(defaultBlock, opts)

	return block
}

func (t *Template) Header(options ...Options) *Block {
	var opts Options
	if len(options) > 0 {
		opts = options[0]
	}

	block := t.newBlock(headerBlock, opts)

	return block
}

func (t *Template) EndHeader(options ...Options) *Block {
	var opts Options
	if len(options) > 0 {
		opts = options[0]
	}

	block := t.newBlock(headerStopBlock, opts)

	return block
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

func Repeat[T any](tmpl *Template, items []T, blockFunc func(b *Block, item T)) {
	for i := range items {
		b := tmpl.Block()
		blockFunc(b, items[i])
	}
}

func (b *Block) height() meter.MM {
	sh := b.slotV2.height()

	if sh > b.h {
		b.h = sh
	}

	return b.h
}

func (b *Block) width() meter.MM {
	sw := b.slotV2.width()

	b.w = sw + b.opts.Spacing

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

func (s *Slot) height() meter.MM {
	s.h = s.printer.height()

	return s.h
}

func (s *Slot) width() meter.MM {
	s.w = s.printer.width()

	return s.w
}

func spacing(sp meter.MM, arrLen int) meter.MM {
	if arrLen < 2 {
		return meter.MM(0)
	}

	return sp * meter.MM(arrLen-1)
}
