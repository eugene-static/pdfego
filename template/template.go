package template

import (
	"github.com/eugene-static/pdf-craft/buffer"
	"github.com/eugene-static/pdf-craft/core"
	"github.com/eugene-static/pdf-craft/meter"
)

const (
	defaultBlock = iota
	headerBlock
	headerStopBlock
	repeatableBlock
)

type Template struct {
	core    *core.Core
	buf     *buffer.Buffer
	page    core.Page
	block   Block
	ordered Ordered
	x0      meter.MM
	y0      meter.MM
	x       meter.MM
	y       meter.MM
}

func New(core *core.Core) *Template {
	core.StartDocument()
	buf := core.AddPage()
	page := core.Page()
	x0, y0 := page.X0Y0()

	return &Template{
		core: core,
		buf:  buf,
		page: page,
		x0:   x0,
		y0:   y0,
		x:    x0,
		y:    y0,
	}
}

func (t *Template) Render() {
	t.render()
	t.core.FinishDocument()
}

type Block struct {
	profile byte
	core    *core.Core
	slots   []Slot
	opts    Options
}

type Slot struct {
	core   *core.Core
	blocks []Block
	table  *Table
	opts   Options
}

type Options struct {
	Spacing meter.MM
	Indent  meter.MM
	Ledge   meter.MM
}

type Ordered interface {
	Len() int
	OrderedRow(int) []string
}

func (t *Template) Bytes() []byte {
	return t.core.Bytes()
}

func (t *Template) Block(opts ...Options) *Block {
	t.newBlock(defaultBlock, opts)

	return &t.block
}

func (t *Template) Header(opts ...Options) *Block {
	t.newBlock(headerBlock, opts)

	return &t.block
}

func (t *Template) EndHeader(opts ...Options) *Block {
	t.newBlock(headerStopBlock, opts)

	return &t.block
}

func (t *Template) Repeater(ordered Ordered) *Template {
	t.ordered = ordered
	t.newBlock(repeatableBlock, nil)

	return t
}

func (t *Template) Repeat(blockFunc func(*Block, []string)) {
	ordered := t.ordered.OrderedRow(0)

	blockFunc(&t.block, ordered)

	for i := range t.ordered.Len() - 1 {
		t.render()

		ordered = t.ordered.OrderedRow(i + 1)

		t.block.update(ordered)
	}
}

func (t *Template) newBlock(profile byte, opts []Options) *Block {
	var opt Options
	if len(opts) != 0 {
		opt = opts[0]
	}

	if t.block.core == nil {
		t.block.core = t.core
		t.block.profile = profile
		t.block.opts = opt

		return &t.block
	}

	t.render()

	t.block.slots = t.block.slots[:0]
	t.block.profile = profile
	t.block.opts = opt

	return &t.block
}

func (b *Block) Slot(opts ...Options) *Slot {
	var opt Options
	if len(opts) != 0 {
		opt = opts[0]
	}

	s := Slot{
		core: b.core,
		opts: opt,
	}

	b.slots = append(b.slots, s)

	return &b.slots[len(b.slots)-1]
}

func (b *Block) Add(blockFunc func(b *Block)) *Block {
	blockFunc(b)

	return b
}

func (s *Slot) Block(opts ...Options) *Block {
	var opt Options
	if len(opts) != 0 {
		opt = opts[0]
	}

	b := Block{
		core: s.core,
		opts: opt,
	}

	s.blocks = append(s.blocks, b)

	return &s.blocks[len(s.blocks)-1]
}

func (s *Slot) Table(rowsNum int, columns []meter.MM) *Table {
	columnsNum := len(columns)

	t := &Table{
		core:      s.core,
		columns:   columns,
		rows:      make([]Row, rowsNum),
		rowspans:  make([]uint8, columnsNum),
		cellsPool: make([]cell, rowsNum*columnsNum),
	}

	s.table = t

	return t
}

func (s *Slot) Add(slotFunc func(s *Slot)) *Slot {
	slotFunc(s)

	return s
}

func (t *Template) render() {
	height := t.block.height()

	if t.block.profile == headerBlock {
		header := t.core.AddHeader()

		t.block.render(header, t.x0, t.y0)
		t.y0 += height
	}

	if t.block.profile == headerStopBlock {
		_, t.y0 = t.page.X0Y0()
	}

	if t.y+height+t.page.Margin() > 0 {
		t.core.WritePage()
		t.core.AddPage()
		t.x, t.y = t.x0, t.y0
	}

	t.block.render(t.buf, t.x, t.y)

	t.y += height
}

func (b *Block) render(buf *buffer.Buffer, x, y meter.MM) {
	y += b.opts.Indent

	for i := range b.slots {
		b.slots[i].render(buf, x, y)

		x += b.slots[i].width() + b.opts.Spacing
	}
}

func (s *Slot) render(buf *buffer.Buffer, x, y meter.MM) {
	if len(s.blocks) > 0 {
		for i := range s.blocks {
			s.blocks[i].render(buf, x, y)

			y += s.blocks[i].height()
		}

		return
	}

	s.table.render(buf, x, y)
}

func (t *Table) render(buf *buffer.Buffer, x, y meter.MM) {
	cx := x
	cy := y

	for i := range t.cellsPool {
		ri := i / len(t.columns)
		ci := i % len(t.columns)

		if t.cellsPool[i].busy {
			t.cellsPool[i].height = t.cellHeight(uint8(ri), t.cellsPool[i].rowspan)

			t.cellsPool[i].render(buf, cx, cy)
		}

		cx += t.columns[ci]

		if ci == len(t.columns)-1 {
			cx = x
			cy += t.rows[ri].height
		}
	}
}

func (b *Block) update(ordered []string) {
	for i := range b.slots {
		b.slots[i].update(ordered)
	}
}

func (s *Slot) update(ordered []string) {
	if s.table != nil {
		s.table.update(ordered)
	}

	for i := range s.blocks {
		s.blocks[i].update(ordered)
	}
}

func (t *Table) update(ordered []string) {
	for i := range t.cellsPool {
		ri := i / len(t.columns)
		ci := i % len(t.columns)

		if ci == 0 {
			t.rows[ri].height = 0
		}

		id := t.cellsPool[i].id
		if int(id) > len(ordered)-1 {
			continue
		}

		t.rows[ri].updateCell(&t.cellsPool[i], ordered[id])
	}
}

func (b *Block) height() (h meter.MM) {
	for i := range b.slots {
		sh := b.slots[i].height()
		if h < sh {
			h = sh
		}
	}

	return h + b.opts.Indent + b.opts.Ledge
}

func (s *Slot) height() (h meter.MM) {
	h = s.opts.Indent + s.opts.Ledge

	if len(s.blocks) > 0 {
		for i := range s.blocks {
			h += s.blocks[i].height()
		}

		return h
	}

	h = s.table.height()

	return h
}

func (b *Block) width() (w meter.MM) {
	for i := range b.slots {
		w += b.slots[i].width()
	}

	w += b.opts.Spacing * meter.MM(len(b.slots)-1)

	return w
}

func (s *Slot) width() (w meter.MM) {
	if len(s.blocks) > 0 {
		for i := range s.blocks {
			bw := s.blocks[i].width()
			if w < bw {
				w = bw
			}
		}

		return w
	}

	w = s.table.width()

	return w
}
