package pdf_craft

import (
	"github.com/eugene-static/pdf-craft/internal/buffer"
	"github.com/eugene-static/pdf-craft/internal/core"
	"github.com/eugene-static/pdf-craft/pkg/unit"
)

const (
	defaultBlock = iota
	headerBlock
	headerStopBlock
	repeatableBlock
)

type Constructor struct {
	core    *core.Core
	buf     *buffer.Buffer
	page    core.Page
	block   Block
	ordered Ordered
	x0      unit.MM
	y0      unit.MM
	x       unit.MM
	y       unit.MM
}

func New(core *core.Core) *Constructor {
	core.StartDocument()
	page := core.Page()
	x0, y0 := page.X0Y0()

	page.New() // ошибки не будет, так как буфер хэдера еще не записан.

	return &Constructor{
		core: core,
		page: page,
		x0:   x0,
		y0:   y0,
		x:    x0,
		y:    y0,
	}
}

func (c *Constructor) Render() {
	c.render()
	c.core.FinishDocument()
}

type Block struct {
	profile byte
	core    *core.Core
	slots   []Slot
	opts    NodeOptions
}

type Slot struct {
	core   *core.Core
	blocks []Block
	table  *Table
	opts   NodeOptions
}

type Ordered interface {
	Len() int
	OrderedRow(int) []string
}

func (c *Constructor) Bytes() ([]byte, error) {
	return c.core.Bytes(), c.core.Err()
}

func (c *Constructor) Block(opts ...NodeOptions) *Block {
	c.newBlock(defaultBlock, opts)

	return &c.block
}

func (c *Constructor) Header(opts ...NodeOptions) *Block {
	c.newBlock(headerBlock, opts)

	return &c.block
}

func (c *Constructor) EndHeader(opts ...NodeOptions) *Block {
	c.newBlock(headerStopBlock, opts)

	return &c.block
}

func (c *Constructor) Repeater(ordered Ordered) *Constructor {
	c.ordered = ordered
	c.newBlock(repeatableBlock, nil)

	return c
}

func (c *Constructor) Add(blockFunc func(*Block, []string)) {
	ordered := c.ordered.OrderedRow(0)

	blockFunc(&c.block, ordered)

	for i := range c.ordered.Len() - 1 {
		c.render()

		ordered = c.ordered.OrderedRow(i + 1)

		c.block.update(ordered)
	}
}

func (c *Constructor) newBlock(profile byte, opts []NodeOptions) *Block {
	options := getOptions(opts)

	if c.block.core == nil {
		c.block.core = c.core
		c.block.profile = profile
		c.block.opts = options

		return &c.block
	}

	c.render()

	c.block.slots = c.block.slots[:0]
	c.block.profile = profile
	c.block.opts = options

	return &c.block
}

func (b *Block) Slot(opts ...NodeOptions) *Slot {
	options := getOptions(opts)

	s := Slot{
		core: b.core,
		opts: options,
	}

	b.slots = append(b.slots, s)

	return &b.slots[len(b.slots)-1]
}

func (b *Block) Add(blockFunc func(b *Block)) *Block {
	blockFunc(b)

	return b
}

func (s *Slot) Block(opts ...NodeOptions) *Block {
	options := getOptions(opts)

	b := Block{
		core: s.core,
		opts: options,
	}

	s.blocks = append(s.blocks, b)

	return &s.blocks[len(s.blocks)-1]
}

func (s *Slot) Table(rowsNum int, columns []unit.MM) *Table {
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

func (c *Constructor) render() {
	height := c.block.height()

	if c.block.profile == headerBlock {
		header := c.page.AddHeader()

		c.block.render(header, c.x0, c.y0)
		c.y0 += height
	}

	if c.block.profile == headerStopBlock {
		_, c.y0 = c.page.X0Y0()
	}

	if c.page.IsBelowBottomBorder(c.y + c.block.opts.IndentTop + height) {
		c.core.RenderPage()

		err := c.page.New()
		if err != nil {
			c.core.WriteError(err)

			return
		}

		c.x, c.y = c.x0, c.y0
	}

	c.block.render(c.page.Buffer(), c.x, c.y)

	c.y += height
}

func (b *Block) render(buf *buffer.Buffer, x, y unit.MM) {
	x += b.opts.IndentLeft
	y += b.opts.IndentTop

	for i := range b.slots {
		b.slots[i].render(buf, x, y)

		x += b.slots[i].width() + b.opts.Spacing
	}
}

func (s *Slot) render(buf *buffer.Buffer, x, y unit.MM) {
	x += s.opts.IndentLeft
	y += s.opts.IndentTop

	if len(s.blocks) > 0 {
		if s.opts.Border != "" {
			borderMask := parseBorder(s.opts.Border)
			width := s.width()
			height := s.height() + s.opts.IndentTop

			borderSize := s.core.DefaultBorderSize()
			if s.opts.BorderSize > 0 {
				borderSize = s.opts.BorderSize
			}

			renderBorder(buf, x, y, width, height, borderMask, borderSize)
		}

		for i := range s.blocks {
			s.blocks[i].render(buf, x, y)

			y += s.blocks[i].height()
		}

		return
	}

	s.table.render(buf, x, y)
}

func (t *Table) render(buf *buffer.Buffer, x, y unit.MM) {
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

func (b *Block) height() (h unit.MM) {
	for i := range b.slots {
		sh := b.slots[i].height()
		if h < sh {
			h = sh
		}
	}

	return h + b.opts.Ledge
}

func (s *Slot) height() (h unit.MM) {
	h = s.opts.Ledge

	if len(s.blocks) > 0 {
		for i := range s.blocks {
			h += s.blocks[i].height() + s.blocks[i].opts.IndentTop + s.blocks[i].opts.Ledge
		}

		return h
	}

	h = s.table.height()

	return h
}

func (b *Block) width() (w unit.MM) {
	for i := range b.slots {
		w += b.slots[i].width() + b.slots[i].opts.IndentLeft
	}

	w += b.opts.Spacing * unit.MM(len(b.slots)-1)

	return w
}

func (s *Slot) width() (w unit.MM) {
	if len(s.blocks) > 0 {
		for i := range s.blocks {
			bw := s.blocks[i].width() + s.blocks[i].opts.IndentLeft
			if w < bw {
				w = bw
			}
		}

		return w
	}

	w = s.table.width()

	return w
}
