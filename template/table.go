package template

import (
	"log/slog"
	"slices"

	"github.com/eugene-static/pdf-craft/buffer"
	"github.com/eugene-static/pdf-craft/core"
	"github.com/eugene-static/pdf-craft/font"
	"github.com/eugene-static/pdf-craft/meter"
)

const (
	borderLeft   uint8                                                 = 1 << iota // 0000 0001 (1)
	borderRight                                                                    // 0000 0010 (2)
	borderTop                                                                      // 0000 0100 (4)
	borderBottom                                                                   // 0000 1000 (8)
	thickLeft                                                                      // 0001 0000 (16)
	thickRight                                                                     // 0010 0000 (32)
	thickTop                                                                       // 0100 0000 (64)
	thickBottom                                                                    // 1000 0000 (128)
	borderAll    = borderLeft | borderRight | borderTop | borderBottom             // 0000 1111 (15)
	thickAll     = thickLeft | thickRight | thickTop | thickBottom                 // 1111 0000 (240)
)

const (
	alignL uint8 = iota
	alignC
	alignR
	alignT
	alignM
	alignB
)

type Table struct {
	core      *core.Core
	columns   []meter.MM
	rows      []Row
	cellsPool []cell
	segBuffer []font.Segment
	rowspans  []uint8
	rowIndex  uint8
	w         meter.MM
	h         meter.MM
	opts      Options
}

func (t *Table) Row() *Row {
	if int(t.rowIndex) > len(t.rows) {
		return &Row{}
	}

	r := &t.rows[t.rowIndex]

	t.newRow(r)

	t.rowIndex++

	return r
}

func (t *Table) Add(rowFunc func(t *Table)) {
	rowFunc(t)
}

func (t *Table) newRow(row *Row) {
	columnsLen := len(t.columns)
	rowspans := t.rowspans

	if t.rowIndex > 0 {
		t.rows[t.rowIndex-1].decrementRowSpans()
	}

	row.core = t.core
	row.height = meter.FontHeight(t.core.DefaultFontSize())
	row.columns = t.columns
	row.columnsLen = uint8(columnsLen)
	row.cells = t.cellsPool[int(t.rowIndex)*columnsLen : int(t.rowIndex)*columnsLen+columnsLen]
	row.rowspans = rowspans
	row.segBuffer = t.segBuffer
}

func (t *Table) height() meter.MM {
	if t.h == 0 {
		for i := range t.rows {
			t.h += t.rows[i].height
		}

		t.h += t.opts.Indent
	}

	return t.h
}

func (t *Table) width() meter.MM {
	if t.w == 0 {
		for i := range t.columns {
			t.w += t.columns[i]
		}
	}

	return t.w
}

func (t *Table) options() Options {
	return t.opts
}

func (t *Table) cellHeight(ri, rowspan uint8) (h meter.MM) {
	for i := ri; i < uint8(len(t.rows)) && i < ri+rowspan; i++ {
		h += t.rows[i].height
	}

	return h
}

type Row struct {
	core        *core.Core
	cells       []cell
	columns     []meter.MM
	rowspans    []uint8
	segBuffer   []font.Segment
	height      meter.MM
	columnIndex uint8
	columnsLen  uint8
}

func (r *Row) CellWithOpts(text string, opts *CellOpts) *Row {
	r.setColIndex()

	if r.columnIndex > r.columnsLen {
		//TODO: error to prevent panic
	}

	c := &r.cells[r.columnIndex]

	r.newCell(c, text, opts)

	linesLen := 1
	if len(c.textWrapped) > 1 {
		linesLen = len(c.textWrapped)
	}

	calcHeight := meter.FontHeight(c.fontSize) * meter.MM(linesLen)

	r.setHeight(c.height, calcHeight)

	r.updateIndexes(c.colspan, c.rowspan)

	return r
}

func (r *Row) Cell(text string) *Row {
	return r.CellWithOpts(text, nil)
}

func (r *Row) Label(text string) *Row {
	return r.CellWithOpts(text, &CellOpts{Align: "LB"})
}

func (r *Row) LabelHead(text string) *Row {
	return r.CellWithOpts(text, &CellOpts{Align: "LB", Font: core.FontBold})
}

func (r *Row) FormL(text string, wrapText bool) *Row {
	return r.CellWithOpts(text, &CellOpts{Align: "LB", Border: "b", Wrap: wrapText})
}

func (r *Row) FormC(text string, wrapText bool) *Row {
	return r.CellWithOpts(text, &CellOpts{Align: "CB", Border: "b", Wrap: wrapText})
}

func (r *Row) Paragraph(text string) *Row {
	return r.CellWithOpts(text, &CellOpts{Align: "CB"})
}

func (r *Row) Underscore(text string) *Row {
	return r.CellWithOpts(text, &CellOpts{Align: "CT", FontSize: r.core.DefaultFontSize().Sub(1)})
}

func (r *Row) Bounded(text string) *Row {
	return r.CellWithOpts(text, &CellOpts{Align: "CM", Border: "o", Wrap: true})
}

func (r *Row) Debug() {
	text := make([][]string, 0, len(r.cells))
	for _, c := range r.cells {
		text = append(text, c.textSegments())
	}

	r.core.Log().Debug("row debug",
		slog.Any("text", text),
		slog.Any("height", r.height),
		slog.Any("width", r.width()),
		slog.Any("rowspans", r.rowspans),
	)
}

func (r *Row) newCell(c *cell, text string, opts *CellOpts) {
	c.font = r.core.Font(core.FontRegular)
	c.fontSize = r.core.DefaultFontSize()
	c.borderSize = r.core.BorderThin()
	c.alignH = alignC
	c.alignV = alignM
	c.colspan = 1
	c.rowspan = 1
	c.busy = true

	c.setOpts(r.core, opts)

	c.width = r.cellWidth(c.colspan)

	if c.wrapped {
		c.textWrapped = c.font.SplitText(text, c.fontSize, c.width, r.segBuffer)
	} else {
		c.textWrapped = c.textSpace[:1]
		c.textWrapped[0] = c.font.FullText(text, c.fontSize)
	}
}

func (r *Row) width() (w float64) {
	for i := range r.cells {
		w += r.cells[i].width.Float64()
	}

	return w
}

// Поле cell.height должно быть равно высоте строки, но пока все ячейки не будут созданы, мы не знаем итоговую высоту строки.
// Поэтому высота ячейки будет определяться в методе render().
func (r *Row) setHeight(optsHeight, calcHeight meter.MM) {
	if calcHeight > r.height {
		r.height = calcHeight
	}

	if optsHeight > 0 {
		r.height = optsHeight
	}
}

// Ширина ячейки равна сумме ширин всех колонок, которые она занимает.
func (r *Row) cellWidth(colspan uint8) (w meter.MM) {
	for i := r.columnIndex; i < r.columnsLen && i < r.columnIndex+colspan; i++ {
		w += r.columns[i]
	}

	return w
}

// Если ячейки предыдущей строки имели rowspan, то мы ищем первую ячейку, которая rowspan не имела, и сдвигаем курсор на неё.
func (r *Row) setColIndex() {
	for r.columnIndex < r.columnsLen && r.rowspans[r.columnIndex] > 0 {
		r.columnIndex++
	}
}

// Каждая ячейка имеет свой colspan > 0. Если colspan > 1, то пропущенным ячейкам тоже необходимо присвоить rowspan этой ячейки.
// Сдвигаем курсор к следующей ячейке.
func (r *Row) updateIndexes(colspan, rowspan uint8) {
	shift := r.columnIndex + colspan

	for i := r.columnIndex; i < min(shift, r.columnsLen); i++ {
		r.rowspans[i] += rowspan
		r.columnIndex = i
	}
}

// При создании новой строки уменьшаем все rowspan на 1.
func (r *Row) decrementRowSpans() {
	for i := range r.rowspans {
		if r.rowspans[i] > 0 {
			r.rowspans[i]--
		}
	}
}

type cell struct {
	font        *font.Font
	textWrapped []font.Segment
	textSpace   [1]font.Segment
	x           meter.MM
	y           meter.MM
	width       meter.MM
	height      meter.MM
	fontSize    meter.PT
	borderSize  meter.PT
	colspan     uint8
	rowspan     uint8
	borderMask  uint8
	alignH      uint8
	alignV      uint8
	wrapped     bool
	busy        bool
}

type CellOpts struct {
	Height     meter.MM
	Colspan    uint8
	Rowspan    uint8
	Align      string
	Border     string
	BorderSize meter.PT
	Font       string
	FontSize   meter.PT
	Wrap       bool
}

func (c *cell) setOpts(core *core.Core, opts *CellOpts) {
	if opts == nil {
		return
	}

	c.height = opts.Height
	c.wrapped = opts.Wrap

	if opts.Font != "" {
		c.font = core.Font(opts.Font)
	}

	if opts.FontSize > 0 {
		c.fontSize = opts.FontSize
	}

	if opts.BorderSize > 0 {
		c.borderSize = opts.BorderSize
	}

	if opts.Colspan > 1 {
		c.colspan = opts.Colspan
	}

	if opts.Rowspan > 1 {
		c.rowspan = opts.Rowspan
	}

	if opts.Align != "" {
		for _, alg := range opts.Align {
			switch alg {
			case 'L':
				c.alignH = alignL
			case 'C':
				c.alignH = alignC
			case 'R':
				c.alignH = alignR
			case 'T':
				c.alignV = alignT
			case 'M':
				c.alignV = alignM
			case 'B':
				c.alignV = alignB
			default:
				//c.alignV = alignM
			}
		}
	}

	if opts.Border != "" {
		for _, b := range opts.Border {
			switch b {
			case 'o':
				c.borderMask |= borderAll
			case 'O':
				c.borderMask |= borderAll | thickAll
			case 'l':
				c.borderMask |= borderLeft
			case 'L':
				c.borderMask |= borderLeft | thickLeft
			case 'r':
				c.borderMask |= borderRight
			case 'R':
				c.borderMask |= borderRight | thickRight
			case 't':
				c.borderMask |= borderTop
			case 'T':
				c.borderMask |= borderTop | thickTop
			case 'b':
				c.borderMask |= borderBottom
			case 'B':
				c.borderMask |= borderBottom | thickBottom
			}
		}
	}
}

func (c *cell) textSegments() []string {
	return slices.Collect(func(yield func(text string) bool) {
		for _, t := range c.textWrapped {
			yield(t.Text())
		}
	})
}

// BT /[FontAlias] [FontSize] Tf 1 0 0 1 [X] [Y] Tm <[TextHex]> Tj ET
func (c *cell) render(buf *buffer.Buffer) {
	if c.borderMask != 0 {
		c.renderBorder(buf)
	}

	if len(c.textWrapped) == 0 {
		return
	}

	buf.WriteStringLn("BT")
	buf.WriteFont(c.font.Alias(), c.fontSize)

	for i := range c.textWrapped {
		dx := c.textDx(c.font, c.textWrapped[i])

		var lenLines = meter.MM(1)

		if len(c.textWrapped) > 1 {
			lenLines = meter.MM(len(c.textWrapped))
		}

		dy := c.textDy(i, lenLines)

		x := c.x + dx
		y := c.y + dy

		buf.WriteText(c.font, x, y, c.textWrapped[i].Text())
	}

	buf.WriteStringLn("ET")
}

func (c *cell) renderBorder(buf *buffer.Buffer) {
	var x0, y0, x1, y1 meter.MM

	if c.borderMask&borderAll == borderAll && (c.borderMask^borderAll == thickAll || c.borderMask^borderAll == 0) {
		bs := c.borderSize
		if c.borderMask&thickAll == thickAll {
			bs *= 4
		}

		buf.WriteRect(bs, c.x, c.y, c.width, c.height)

		return
	}

	if c.borderMask&borderLeft != 0 {
		bs := c.borderSize
		if c.borderMask&thickLeft != 0 {
			bs *= 4
		}

		x0 = c.x
		y0 = c.y
		x1 = x0
		y1 = c.y + c.height

		buf.WriteLine(bs, x0, y0, x1, y1)
	}

	if c.borderMask&borderRight != 0 {
		bs := c.borderSize
		if c.borderMask&thickRight != 0 {
			bs *= 4
		}

		x0 = c.x + c.width
		y0 = c.y
		x1 = x0
		y1 = c.y + c.height

		buf.WriteLine(bs, x0, y0, x1, y1)
	}

	if c.borderMask&borderTop != 0 {
		bs := c.borderSize
		if c.borderMask&thickTop != 0 {
			bs *= 4
		}

		x0 = c.x
		y0 = c.y
		x1 = c.x + c.width
		y1 = c.y

		buf.WriteLine(bs, x0, y0, x1, y1)
	}

	if c.borderMask&borderBottom != 0 {
		bs := c.borderSize
		if c.borderMask&thickBottom != 0 {
			bs *= 4
		}

		x0 = c.x
		y0 = c.y + c.height
		x1 = c.x + c.width
		y1 = y0

		buf.WriteLine(bs, x0, y0, x1, y1)
	}
}

func (c *cell) textDx(f *font.Font, seg font.Segment) (dx meter.MM) {
	textWidth := seg.Width()

	if textWidth == 0 {
		text := []rune(seg.Text())
		textWidth = f.MeasureText(c.fontSize, text, 0, len(text)).MM()
	}

	switch c.alignH {
	case alignR:
		dx = c.width - textWidth - 0.2
	case alignC:
		dx = (c.width - textWidth) / 2
	default:
		// чтобы текст не прилипал к границе
		dx = 0.2
	}

	return dx
}

func (c *cell) textDy(index int, lenLines meter.MM) (dy meter.MM) {
	fontHeight := c.fontSize.MM()

	k := meter.MM(0.3)

	// index + 1 необходим для того, чтобы выставить Y-координату по верхнему краю шрифта.
	// PDF считает Y от нижней границы страницы, а здесь все координаты указаны от верхней. К тому же позиционирует шрифт по baseline.
	// Поэтому для верного позиционирования шрифта нам необходимо добавить еще одну высоту строки.
	baseLineDy := meter.MM(index+1) * fontHeight

	switch c.alignV {
	case alignB:
		dy = c.height - lenLines*fontHeight - k
	case alignM:
		dy = (c.height - lenLines*fontHeight) / 2
	default:
		dy = k
	}

	dy += baseLineDy

	return dy
}
