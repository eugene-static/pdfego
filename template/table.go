package template

import (
	"log/slog"
	"slices"
	"strings"
	"unicode"

	"github.com/eugene-static/pdf-craft/buffer"
	"github.com/eugene-static/pdf-craft/core"
	"github.com/eugene-static/pdf-craft/font"
	"github.com/eugene-static/pdf-craft/meter"
)

type Table struct {
	core    *core.Core
	columns []meter.MM
	rows    []*Row
	opts    Options
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
		buffer:     make([]font.Segment, 0, 20),
	}

	return r
}

func (t *Table) render(buf *buffer.Buffer, x, y meter.MM) {
	height := meter.MM(0.0)

	for rowIndex, row := range t.rows {
		y0 := y + height
		height += row.height

		cellX := x

		for i, c := range row.cells {
			c.x = cellX
			c.y = y0
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

type Row struct {
	core        *core.Core
	height      meter.MM
	cells       []cell
	rowspans    []int
	columns     []meter.MM
	buffer      []font.Segment
	columnIndex int
	columnsLen  int
}

func (r *Row) Cell(text string, opts ...CellOpts) *Row {
	var opt CellOpts

	if opts != nil {
		opt = opts[0]
	}

	r.setColIndex()

	if r.columnIndex > r.columnsLen {
		//TODO: error to prevent panic
	}

	c := r.newCell(text, opt)

	r.cells[r.columnIndex] = c

	calcHeight := meter.FontHeight(c.fontSize) * meter.MM(len(c.text))

	r.setHeight(c.height, calcHeight)

	r.updateIndexes(c.colspan, c.rowspan)

	return r
}

func (r *Row) Label(text string) *Row {
	return r.Cell(text, CellOpts{Align: "LB"})
}

func (r *Row) LabelHead(text string) *Row {
	return r.Cell(text, CellOpts{Align: "LB", Font: core.FontBold})
}

func (r *Row) FormL(text string, wrapText bool) *Row {
	return r.Cell(text, CellOpts{Align: "LB", Border: "b", Wrap: wrapText})
}

func (r *Row) FormC(text string, wrapText bool) *Row {
	return r.Cell(text, CellOpts{Align: "CB", Border: "b", Wrap: wrapText})
}

func (r *Row) Paragraph(text string) *Row {
	return r.Cell(text, CellOpts{Align: "CB"})
}

func (r *Row) Underscore(text string) *Row {
	return r.Cell(text, CellOpts{Align: "CT", FontSize: r.core.DefaultFontSize().Sub(1)})
}

func (r *Row) Bounded(text string) *Row {
	return r.Cell(text, CellOpts{Align: "CM", Border: "o", Wrap: true})
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

func (r *Row) newCell(text string, opts CellOpts) cell {
	c := cell{
		core:       r.core,
		height:     opts.Height,
		text:       []font.Segment{font.NewSegment(text, 0)},
		font:       r.core.Font(core.FontRegular),
		fontSize:   r.core.DefaultFontSize(),
		border:     opts.Border,
		borderSize: r.core.BorderThin(),
		align:      "CM",
		colspan:    1,
		rowspan:    1,
		busy:       true,
	}

	if opts.Font != "" {
		c.font = r.core.Font(strings.ToUpper(opts.Font))
	}

	if opts.FontSize > 0 {
		c.fontSize = opts.FontSize
	}

	if opts.BorderSize > 0 {
		c.borderSize = meter.PT(opts.BorderSize)
	}

	if opts.Align != "" {
		c.align = opts.Align
	}

	if opts.Colspan > 1 {
		c.colspan = opts.Colspan
	}

	if opts.Rowspan > 1 {
		c.rowspan = opts.Rowspan
	}

	c.width = r.cellWidth(c.colspan)

	c.font.SaveRunes(text)

	if opts.Wrap {
		split := c.font.SplitTextOptimized(r.buffer, text, c.fontSize, c.width)

		c.text = split
	}

	return c
}

func (r *Row) width() (w float64) {
	for _, c := range r.cells {
		w += c.width.Float64()
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
func (r *Row) cellWidth(colspan int) (w meter.MM) {
	for i := r.columnIndex; i < min(r.columnsLen, r.columnIndex+colspan); i++ {
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
func (r *Row) updateIndexes(colspan, rowspan int) {
	shift := r.columnIndex + colspan

	for i := r.columnIndex; i < min(shift, r.columnsLen); i++ {
		r.rowspans[i] += rowspan
		r.columnIndex = i
	}
}

// При создании новой строки уменьшаем все rowspan на 1.
func (r *Row) decrementRowSpans() []int {
	for i := range r.rowspans {
		r.rowspans[i]--
	}

	return r.rowspans
}

type cell struct {
	core       *core.Core
	x          meter.MM
	y          meter.MM
	width      meter.MM
	height     meter.MM
	colspan    int
	rowspan    int
	font       *font.Font
	fontSize   meter.PT
	border     string
	borderSize meter.PT
	align      string
	text       []font.Segment
	busy       bool
}

type CellOpts struct {
	Height     meter.MM
	Colspan    int
	Rowspan    int
	Align      string
	Border     string
	BorderSize float64
	Font       string
	FontSize   meter.PT
	Wrap       bool
}

func (c *cell) textSegments() []string {
	return slices.Collect(func(yield func(text string) bool) {
		for _, t := range c.text {
			yield(t.Text())
		}
	})
}

// BT /[FontAlias] [FontSize] Tf 1 0 0 1 [X] [Y] Tm <[TextHex]> Tj ET
func (c *cell) render(buf *buffer.Buffer) {
	if len(c.text) == 0 {
		return
	}

	buf.WriteStringLn("BT")
	buf.WriteFont(c.font.Alias(), c.fontSize)

	for i, seg := range c.text {
		dx := c.textDx(c.font, seg)
		dy := c.textDy(i)

		x := c.x + dx
		y := c.y + dy

		buf.WriteText(c.font, x, y, seg.Text())
	}

	buf.WriteStringLn("ET")

	if c.border != "" {
		c.renderBorder(buf)
	}
}

func (c *cell) renderBorder(buf *buffer.Buffer) {
	bs := c.borderSize

	for _, b := range c.border {
		if unicode.IsUpper(b) {
			bs = c.core.BorderThick()
		}

		var x0, y0, x1, y1 meter.MM

		switch b {
		case 'o', 'O':
			buf.WriteRect(bs, c.x, c.y, c.width, c.height)

			return
		case 't', 'T':
			x0 = c.x
			y0 = c.y
			x1 = c.x + c.width
			y1 = c.y
		case 'r', 'R':
			x0 = c.x + c.width
			y0 = c.y
			x1 = x0
			y1 = c.y + c.height
		case 'b', 'B':
			x0 = c.x
			y0 = c.y + c.height
			x1 = c.x + c.width
			y1 = y0
		case 'l', 'L':
			x0 = c.x
			y0 = c.y
			x1 = x0
			y1 = c.y + c.height
		default:
			continue
		}

		buf.WriteLine(bs, x0, y0, x1, y1)
	}
}

func (c *cell) textDx(f *font.Font, seg font.Segment) (dx meter.MM) {
	textWidth := seg.Width()

	if textWidth == 0 {
		textWidth = f.MeasureText(c.fontSize, seg.Text()).MM()
	}

	switch {
	case strings.ContainsRune(c.align, 'R'):
		dx = c.width - textWidth - 0.2
	case strings.ContainsRune(c.align, 'C'):
		dx = (c.width - textWidth) / 2
	default:
		// чтобы текст не прилипал к границе
		dx = 0.2
	}

	return dx
}

func (c *cell) textDy(index int) (dy meter.MM) {
	lenLines := meter.MM(len(c.text))
	fontHeight := c.fontSize.MM()

	k := meter.MM(0.3)

	// index + 1 необходим для того, чтобы выставить Y-координату по верхнему краю шрифта.
	// PDF считает Y от нижней границы страницы, а здесь все координаты указаны от верхней. К тому же позиционирует шрифт по baseline.
	// Поэтому для верного позиционирования шрифта нам необходимо добавить еще одну высоту строки.
	baseLineDy := meter.MM(index+1) * fontHeight

	switch {
	case strings.ContainsRune(c.align, 'B'):
		dy = c.height - lenLines*fontHeight - k
	case strings.ContainsRune(c.align, 'M'):
		dy = (c.height - lenLines*fontHeight) / 2
	default:
		dy = k
	}

	dy += baseLineDy

	return dy
}
