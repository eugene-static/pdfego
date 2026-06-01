package template

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/eugene-static/pdf-craft/internal/buffer"
	"github.com/eugene-static/pdf-craft/internal/core/core"
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
	isPrinted bool
	opts      NodeOptions
}

type Row struct {
	core        *core.Core
	cells       []cell
	columns     []unit.MM
	rowspans    []uint8
	height      unit.MM
	columnIndex uint8
	columnsLen  uint8
}

type cell struct {
	id          uint8
	font        *font.Font
	textWrapped []font.Text
	width       unit.MM
	height      unit.MM
	fontSize    unit.PT
	borderSize  unit.PT
	borderMask  uint8
	colspan     uint8
	rowspan     uint8
	alignH      uint8
	alignV      uint8
	wrapped     bool
	static      bool
	busy        bool
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
	//row.height = unit.FontHeight(t.core.DefaultFontSize())
	row.columns = t.columns
	row.columnsLen = uint8(columnsLen)
	row.cells = t.cellsPool[int(t.rowIndex)*columnsLen : int(t.rowIndex)*columnsLen+columnsLen]
	row.rowspans = rowspans
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

func (r *Row) CellWithOpts(text string, opts *CellOptions) *Row {
	r.setColIndex()

	if r.columnIndex > r.columnsLen {
		err := fmt.Errorf("column index %d out of bounds with max = %d", r.columnIndex, r.columnsLen)
		r.core.WriteError(err)

		return r
	}

	c := &r.cells[r.columnIndex]

	r.newCell(c, text, opts)
	r.setHeight(c)
	r.updateIndexes(c)

	return r
}

func (r *Row) Cell(text string) *Row {
	return r.CellWithOpts(text, nil)
}

func (r *Row) Label(text string) *Row {
	return r.CellWithOpts(text, &CellOptions{Align: "LB"})
}

func (r *Row) LabelHead(text string) *Row {
	return r.CellWithOpts(text, &CellOptions{Align: "LB", Font: core.FontBold})
}

func (r *Row) FormL(text string, wrapText bool) *Row {
	return r.CellWithOpts(text, &CellOptions{Align: "LB", Border: "b", Wrap: wrapText})
}

func (r *Row) FormC(text string, wrapText bool) *Row {
	return r.CellWithOpts(text, &CellOptions{Align: "CB", Border: "b", Wrap: wrapText})
}

func (r *Row) Paragraph(text string) *Row {
	return r.CellWithOpts(text, &CellOptions{Align: "CB"})
}

func (r *Row) Underscore(text string) *Row {
	return r.CellWithOpts(text, &CellOptions{Align: "CT", FontSize: r.core.DefaultFontSize().Sub(1)})
}

func (r *Row) Bounded(text string) *Row {
	return r.CellWithOpts(text, &CellOptions{Align: "CM", Border: "o", Wrap: false})
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

func (r *Row) newCell(c *cell, text string, opts *CellOptions) {
	c.textWrapped = make([]font.Text, 0, 10)
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
		c.textWrapped = append(c.textWrapped, c.font.SplitText(text, c.fontSize, c.width, c.textWrapped)...)
	} else {
		c.textWrapped = append(c.textWrapped, c.font.FullText(text, c.fontSize))
	}
}

func (r *Row) width() (w float64) {
	for i := range r.cells {
		w += r.cells[i].width.Float64()
	}

	return w
}

func (r *Row) updateCell(c *cell, text string) {
	c.textWrapped = c.textWrapped[:0]

	if c.wrapped {
		c.textWrapped = append(c.textWrapped, c.font.SplitText(text, c.fontSize, c.width, c.textWrapped)...)
	} else {
		c.textWrapped = append(c.textWrapped, c.font.FullText(text, c.fontSize))
	}

	r.setHeight(c)
}

// Поле cell.height должно быть равно высоте строки, но пока все ячейки не будут созданы, мы не знаем итоговую высоту строки.
// Поэтому высота ячейки будет определяться в методе render().
// TODO: сейчас я не считаю высоту, если есть rowspan, т.к. не знаю высоту следующей строки. Может быть нужно её учитывать
func (r *Row) setHeight(c *cell) {
	//calcHeight := unit.FontHeight(c.fontSize) * unit.MM(len(c.textWrapped))
	calcHeight := c.font.Height(c.fontSize).MM() * unit.MM(len(c.textWrapped))

	if calcHeight > r.height && c.rowspan < 2 {
		r.height = calcHeight
	}

	if c.static && c.height > r.height {
		r.height = c.height
	}
}

// Ширина ячейки равна сумме ширин всех колонок, которые она занимает.
func (r *Row) cellWidth(colspan uint8) (w unit.MM) {
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
func (r *Row) updateIndexes(c *cell) {
	shift := r.columnIndex + c.colspan

	for i := r.columnIndex; i < min(shift, r.columnsLen); i++ {
		r.rowspans[i] += c.rowspan
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

func (c *cell) setOpts(core *core.Core, opts *CellOptions) {
	if opts == nil {
		return
	}

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

	//if opts.Align != "" {
	//	for _, alg := range opts.Align {
	//		switch alg {
	//		case 'L':
	//			c.alignH = alignL
	//		case 'C':
	//			c.alignH = alignC
	//		case 'R':
	//			c.alignH = alignR
	//		case 'T':
	//			c.alignV = alignT
	//		case 'M':
	//			c.alignV = alignM
	//		case 'B':
	//			c.alignV = alignB
	//		default:
	//			//c.alignV = alignM
	//		}
	//	}
	//}
}

func (c *cell) textSegments() []string {
	return slices.Collect(func(yield func(text string) bool) {
		for _, t := range c.textWrapped {
			yield(t.Data())
		}
	})
}

// BT /[FontAlias] [FontSize] Tf 1 0 0 1 [X] [Y] Tm <[TextHex]> Tj ET
func (c *cell) render(buf *buffer.Buffer, x, y unit.MM) {
	if c.borderMask != 0 {
		renderBorder(buf, x, y, c.width, c.height, c.borderMask, c.borderSize)
	}

	if len(c.textWrapped) == 0 {
		return
	}

	buf.WriteStringLn("BT")
	buf.WriteFont(c.font.Alias(), c.fontSize)

	for i := range c.textWrapped {
		dx := c.textDx(c.font, c.textWrapped[i])

		var lenLines = unit.MM(1)

		if len(c.textWrapped) > 1 {
			lenLines = unit.MM(len(c.textWrapped))
		}

		dy := c.textDy(i, lenLines)

		cx := x + dx
		cy := y + dy

		buf.WriteText(c.font, cx, cy, c.textWrapped[i].Data())
	}

	buf.WriteStringLn("ET")
}

func (c *cell) textDx(f *font.Font, seg font.Text) (dx unit.MM) {
	textWidth := seg.Width()

	if textWidth == 0 {
		text := []rune(seg.Data())
		textWidth = f.MeasureText(c.fontSize, text, 0, len(text)).MM()
	}

	switch c.alignH {
	case alignR:
		dx = c.width - textWidth - unit.Padding(c.fontSize)
	case alignC:
		dx = (c.width - textWidth) / 2
	default:
		dx = unit.Padding(c.fontSize)
	}

	return dx
}

func (c *cell) textDy(index int, lenLines unit.MM) (dy unit.MM) {
	fontHeight := c.fontSize.MM()

	// index + 1 необходим для того, чтобы выставить Y-координату по верхнему краю шрифта.
	// PDF считает Y от нижней границы страницы, а здесь все координаты указаны от верхней. К тому же позиционирует шрифт по baseline.
	// Поэтому для верного позиционирования шрифта нам необходимо добавить еще одну высоту строки.
	baseLineDy := unit.MM(index+1) * fontHeight

	switch c.alignV {
	case alignB:
		dy = c.height - lenLines*fontHeight - unit.Padding(c.fontSize)
	case alignM:
		dy = (c.height - lenLines*fontHeight) / 2
	default:
		dy = unit.Padding(c.fontSize)
	}

	dy += baseLineDy

	return dy
}
