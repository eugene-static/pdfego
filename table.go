package pdf_craft

import (
	"log/slog"
	"slices"

	"github.com/eugene-static/pdf-craft/internal/buffer"
	"github.com/eugene-static/pdf-craft/internal/core"
	"github.com/eugene-static/pdf-craft/internal/font"
	"github.com/eugene-static/pdf-craft/pkg/unit"
)

type Table struct {
	core      *core.Core
	columns   []unit.MM
	rows      []Row
	cellsPool []cell
	rowspans  []uint8
	rowIndex  uint8
	opts      NodeOptions
}

func (t *Table) Row(opts ...NodeOptions) *Row {
	options := getOptions(opts)

	t.setRowIndex()

	r := &t.rows[t.rowIndex]

	t.decrementRowSpans()
	t.newRow(r, options)
	t.updateIndexes()

	return r
}

func (t *Table) Add(rowFunc func(t *Table)) {
	rowFunc(t)
}

func (t *Table) newRow(row *Row, options NodeOptions) {
	columnsLen := len(t.columns)
	start := int(t.rowIndex) * columnsLen
	end := start + columnsLen

	row.core = t.core
	row.columns = t.columns
	row.columnsLen = uint8(columnsLen)
	row.cells = t.cellsPool[start:end]
	row.rowspans = t.rowspans
	row.options = options
}

func (t *Table) setRowIndex() {
	if int(t.rowIndex) >= len(t.rows) {
		t.rowIndex--
	}
}

func (t *Table) updateIndexes() {
	t.rowIndex++
}

func (t *Table) height() (h unit.MM) {
	for i := range t.rows {
		h += t.rows[i].height
	}

	h += t.opts.IndentTop

	return h
}

func (t *Table) width() (w unit.MM) {
	for i := range t.columns {
		w += t.columns[i]
	}

	return w
}

func (t *Table) cellHeight(ri, rowspan uint8) (h unit.MM) {
	for i := ri; i < uint8(len(t.rows)) && i < ri+rowspan; i++ {
		h += t.rows[i].height
	}

	return h
}

// При создании новой строки уменьшаем все rowspan на 1.
func (t *Table) decrementRowSpans() {
	for i := range t.rowspans {
		if t.rowspans[i] > 0 {
			t.rowspans[i]--
		}
	}
}

type Row struct {
	core        *core.Core
	cells       []cell
	columns     []unit.MM
	rowspans    []uint8
	height      unit.MM
	columnIndex uint8
	columnsLen  uint8
	options     NodeOptions
}

func (r *Row) Cell(text string, opts ...CellOptions) *Row {
	options := getOptions(opts)

	r.setColIndex()

	c := &r.cells[r.columnIndex]

	r.newCell(c, text, options)
	r.setHeight(c)
	r.updateIndexes(c)

	return r
}

func (r *Row) Label(text string) *Row {
	return r.Cell(text, CellOptions{Align: "LB"})
}

func (r *Row) LabelHead(text string) *Row {
	return r.Cell(text, CellOptions{Align: "LB", Font: core.FontBold})
}

func (r *Row) FormL(text string, wrapText bool) *Row {
	return r.Cell(text, CellOptions{Align: "LB", Border: "b", Wrap: wrapText})
}

func (r *Row) FormC(text string, wrapText bool) *Row {
	return r.Cell(text, CellOptions{Align: "CB", Border: "b", Wrap: wrapText})
}

func (r *Row) Paragraph(text string) *Row {
	return r.Cell(text, CellOptions{Align: "CB"})
}

func (r *Row) Underscore(text string) *Row {
	return r.Cell(text, CellOptions{Align: "CT", FontSize: r.core.DefaultFontSize().Sub(1)})
}

func (r *Row) Bounded(text string) *Row {
	return r.Cell(text, CellOptions{Align: "CM", Border: "o", Wrap: false})
}

func (r *Row) Debug() {
	text := make([][]string, 0, len(r.cells))
	for _, c := range r.cells {
		text = append(text, c.lines())
	}

	r.core.Log().Debug("row debug",
		slog.Any("text", text),
		slog.Any("height", r.height),
		slog.Any("width", r.width()),
		slog.Any("rowspans", r.rowspans),
	)
}

func (r *Row) newCell(c *cell, text string, opts CellOptions) {
	c.textLines = make([]font.Text, 0, 10)
	c.font = r.core.Font(core.FontRegular)
	c.fontSize = r.core.DefaultFontSize()
	c.borderSize = r.core.DefaultBorderSize()
	c.alignH = alignC
	c.alignV = alignM
	c.colspan = 1
	c.rowspan = 1
	c.busy = true

	c.setOpts(r.core, opts)

	c.width = r.cellWidth(c.colspan)

	if c.wrapped {
		c.textLines = append(c.textLines, c.font.SplitText(text, c.fontSize, c.width, c.textLines)...)
	} else {
		c.textLines = append(c.textLines, c.font.FullText(text, c.fontSize))
	}
}

func (r *Row) updateCell(c *cell, text string) {
	c.textLines = c.textLines[:0]

	if c.wrapped {
		c.textLines = append(c.textLines, c.font.SplitText(text, c.fontSize, c.width, c.textLines)...)
	} else {
		c.textLines = append(c.textLines, c.font.FullText(text, c.fontSize))
	}

	r.setHeight(c)
}

func (r *Row) width() (w float64) {
	for i := range r.cells {
		w += r.cells[i].width.Float64()
	}

	return w
}

// Ширина ячейки равна сумме ширин всех колонок, которые она занимает.
func (r *Row) cellWidth(colspan uint8) (w unit.MM) {
	for i := r.columnIndex; i < r.columnsLen && i < r.columnIndex+colspan; i++ {
		w += r.columns[i]
	}

	return w
}

// Поле cell.height должно быть равно высоте строки, но пока все ячейки не будут созданы, мы не знаем итоговую высоту строки.
// Поэтому высота ячейки будет определяться в методе render().
// TODO: сейчас я не считаю высоту, если есть rowspan, т.к. не знаю высоту следующей строки. Может быть нужно её учитывать
func (r *Row) setHeight(c *cell) {
	//calcHeight := unit.FontHeight(c.fontSize) * unit.MM(len(c.textLines))
	calcHeight := c.font.Height(c.fontSize).MM() * unit.MM(len(c.textLines))

	if calcHeight > r.height && c.rowspan < 2 {
		r.height = calcHeight
	}

	if c.static && c.height > r.height {
		r.height = c.height
	}
}

// Если ячейки предыдущей строки имели rowspan, то мы ищем первую ячейку, которая rowspan не имела, и сдвигаем курсор на неё.
func (r *Row) setColIndex() {
	for r.columnIndex < r.columnsLen && r.rowspans[r.columnIndex] > 0 {
		r.columnIndex++
	}
}

// Каждая ячейка имеет свой colspan > 0. Если colspan > 1, то пропущенным ячейкам тоже необходимо присвоить rowspan этой ячейки.
// Сдвигаем курсор к следующей ячейке.
func (r *Row) updateIndexes(c *cell) {
	shift := r.columnIndex + c.colspan

	for i := r.columnIndex; i < min(shift, r.columnsLen); i++ {
		r.rowspans[i] += c.rowspan
		r.columnIndex = i
	}
}

type cell struct {
	id         uint8
	font       *font.Font
	textLines  []font.Text
	width      unit.MM
	height     unit.MM
	fontSize   unit.PT
	borderSize unit.PT
	borderMask uint8
	colspan    uint8
	rowspan    uint8
	alignH     uint8
	alignV     uint8
	wrapped    bool
	static     bool
	busy       bool
}

func (c *cell) setOpts(core *core.Core, opts CellOptions) {
	c.id = opts.ID
	c.wrapped = opts.Wrap
	c.height = opts.Height
	c.static = opts.Height > 0
	c.borderMask = parseBorder(opts.Border)
	c.alignH, c.alignV = parseAlignment(opts.Align)

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
}

func (c *cell) lines() []string {
	return slices.Collect(func(yield func(text string) bool) {
		for _, t := range c.textLines {
			yield(t.Data())
		}
	})
}

// BT /[FontAlias] [FontSize] Tf 1 0 0 1 [X] [Y] Tm <[TextHex]> Tj ET
func (c *cell) render(buf *buffer.Buffer, x, y unit.MM) {
	if c.borderMask != 0 {
		renderBorder(buf, x, y, c.width, c.height, c.borderMask, c.borderSize)
	}

	if len(c.textLines) == 0 {
		return
	}

	buf.WriteStringLn("BT")
	buf.WriteFont(c.font.Alias(), c.fontSize)

	for i := range c.textLines {
		dx := c.lineDx(i)
		dy := c.lineDy(i)

		cx := x + dx
		cy := y + dy

		buf.WriteText(c.font, cx, cy, c.textLines[i].Data())
	}

	buf.WriteStringLn("ET")
}

func (c *cell) lineDx(index int) (dx unit.MM) {
	textWidth := c.textLines[index].Width()

	if textWidth == 0 {
		text := []rune(c.textLines[index].Data())
		textWidth = c.font.MeasureText(c.fontSize, text, 0, len(text)).MM()
	}

	switch c.alignH {
	case alignR:
		dx = c.width - textWidth - font.Padding(c.fontSize).MM()
	case alignC:
		dx = (c.width - textWidth) / 2
	default:
		dx = font.Padding(c.fontSize).MM()
	}

	return dx
}

func (c *cell) lineDy(index int) (dy unit.MM) {
	fontHeight := c.font.Height(c.fontSize).MM()
	lenTextLines := len(c.textLines)
	padding := font.Padding(c.fontSize).MM()

	// index + 1 необходим для того, чтобы выставить Y-координату по верхнему краю шрифта.
	// PDF считает Y от нижней границы страницы, а здесь все координаты указаны от верхней. К тому же позиционирует шрифт по baseline.
	// Поэтому для верного позиционирования шрифта нам необходимо добавить еще одну высоту строки.
	dy = unit.MM(index+1)*fontHeight - padding

	switch c.alignV {
	case alignB:
		dy += c.height - unit.MM(lenTextLines)*fontHeight
	case alignM:
		dy += (c.height - unit.MM(lenTextLines)*fontHeight) / 2
	default:
		dy += 0
	}

	return dy
}
