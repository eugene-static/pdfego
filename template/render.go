package template

import (
	"github.com/eugene-static/pdf-craft/meter"
)

func (t *Template) newBlock(profile byte, opts Options) *Block {
	if t.block == nil {
		t.block = &Block{
			profile: profile,
			core:    t.core,
			buf:     t.buf,
			x:       t.x0,
			y:       t.y0,
			opts:    opts,
		}

		return t.block
	}

	t.print()
	//t.block.slots = t.block.slots[:0]
	t.block.profile = profile
	t.block.opts = opts
	t.block.x = t.x
	t.block.y = t.y

	return t.block
}

func (t *Template) Render() {
	t.print()
	t.core.FinishDocument()
}

func (t *Template) print() {
	height := t.block.height()

	switch t.block.profile {
	case headerBlock:
		headerBuf := t.core.AddHeader()
		t.block.buf = headerBuf

		t.block.print(t.x0, t.y0)

		t.headerHeight = height
	case headerStopBlock:
		t.core.RemoveHeader()

		t.headerHeight = 0
	default:
		//
	}

	if (t.page.Margin() + t.y + height) > 0 {
		t.core.WritePage()
		t.core.AddPage()
		t.x = t.x0
		t.y = t.y0 + t.headerHeight
	}

	t.block.buf = t.buf

	t.block.print(t.x, t.y)

	t.y += height
}

func (b *Block) print(x, y meter.MM) {
	x += b.w
	y += b.opts.Indent
	b.slotV2.print(x, y)
}

func (b *Block) SlotV2() *Slot {
	if b.slotV2 == nil {
		b.slotV2 = &Slot{
			core: b.core,
			buf:  b.buf,
			x:    b.x,
			y:    b.y,
		}

		return b.slotV2
	}

	b.print(b.x, b.y)

	b.w += b.slotV2.width()

	return b.slotV2
}

func (s *Slot) print(x, y meter.MM) {
	s.printer.print(x, y)
}

func (s *Slot) BlockV2(options ...Options) *Block {
	var opts Options
	if len(options) > 0 {
		opts = options[0]
	}

	if s.block == nil {
		s.block = &Block{
			core: s.core,
			opts: opts,
			buf:  s.buf,
			x:    s.x,
			y:    s.y + opts.Indent,
		}

		s.printer = s.block

		return s.block
	}

	s.print(s.x, s.y+opts.Indent)

	s.w = s.printer.width()
	s.h += s.printer.height()
	s.block.opts = opts
	s.block.x = s.x
	s.block.y = s.y + s.h + opts.Indent
	s.block.w = 0
	s.block.h = 0
	s.printer = s.block

	return s.block
}

func (s *Slot) TableV2(rowsNum int, columns []meter.MM) *Table {
	columnsNum := len(columns)

	if s.table == nil {
		s.table = &Table{
			core:      s.core,
			buf:       s.buf,
			x:         s.x,
			y:         s.y,
			columns:   columns,
			rows:      make([]Row, rowsNum),
			rowspans:  make([]uint8, columnsNum),
			cellsPool: make([]cell, rowsNum*columnsNum),
		}

		s.printer = s.table

		return s.table
	}

	s.print(s.x, s.y)

	s.table.rowIndex = 0
	s.table.rows = resize(s.table.rows, rowsNum)
	s.table.cellsPool = resize(s.table.cellsPool, rowsNum*columnsNum)
	s.table.rowspans = resize(s.table.rowspans, columnsNum)
	s.table.columns = resize(s.table.columns, columnsNum)
	copy(s.table.columns, columns)

	s.printer = s.table

	return s.table
}

func (t *Table) print(x, y meter.MM) {
	height := meter.MM(0.0)

	for ri := range t.rows {
		x0 := t.x
		y0 := t.y + height
		height += t.rows[ri].height

		for ci := range t.rows[ri].cells {
			t.rows[ri].cells[ci].x = x0
			t.rows[ri].cells[ci].y = y0
			x0 += t.columns[ci]

			if !t.rows[ri].cells[ci].busy {
				continue
			}

			t.rows[ri].cells[ci].height = t.cellHeight(uint8(ri), t.rows[ri].cells[ci].rowspan)

			t.rows[ri].cells[ci].render(t.buf)
		}
	}
}

func resize[T any](arr []T, size int) []T {
	if size > cap(arr) {
		arr = make([]T, size)

		return arr
	}

	arr = arr[:size]
	clear(arr)

	return arr
}
