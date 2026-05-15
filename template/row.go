package template

import (
	"strings"

	core "github.com/eugene-static/pdf-craft/core"
	"github.com/eugene-static/pdf-craft/meter"
)

type Row struct {
	core        *core.Core
	x           meter.MM
	y           meter.MM
	height      meter.MM
	cells       []cell
	rowspans    []int
	columns     []meter.MM
	columnIndex int
	columnsLen  int
}

func (r *Row) Cell(text string, opts ...CellOpts) {
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

	//r.updateRowSpans(c.rowspan)

	r.updateColIndex(c.colspan)
}

func (r *Row) newCell(text string, opts CellOpts) cell {
	c := cell{
		core:       r.core,
		text:       []string{text},
		font:       core.FontRegular,
		fontSize:   r.core.DefaultFontSize(),
		border:     "",
		borderSize: r.core.BorderThin(),
		align:      "CM",
		colspan:    1,
		rowspan:    1,
		busy:       true,
	}

	//if c.width == -1 {
	//	c.width = c.core.page.width - c.core.page.margin - c.x
	//}

	if opts.Font != "" {
		c.font = strings.ToUpper(opts.Font)
	}

	if opts.FontSize > 0 {
		c.fontSize = meter.PT(opts.FontSize)
	}

	if opts.Border != "" {
		c.border = opts.Border
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

	f := c.core.Font(c.font)

	f.SaveRunes(text)

	if opts.Wrap {
		split := f.SplitText(text, c.fontSize, c.width)

		c.text = split
	}

	height := meter.FontHeight(c.fontSize) * meter.MM(len(c.text))

	r.setHeight(meter.MM(opts.Height), height)

	return c
}

// Поле cell.height должно быть равно высоте строки, но пока все ячейки не будут созданы, мы не знаем итоговую высоту строки.
// Поэтому высота ячейки будет определяться в методе render().
func (r *Row) setHeight(heights ...meter.MM) {
	for _, h := range heights {
		if h > r.height {
			r.height = h
		}
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
func (r *Row) updateColIndex(colspan int) {
	for i := r.columnIndex; i < min(r.columnIndex+colspan, r.columnsLen); i++ {
		r.rowspans[i]++
		r.columnIndex = i
	}
}

// Изначально массив rowspan содержит только 0. При создании ячейки она имеет по-умолчанию rowspan = 1, так как занимает одну строку.
// Записываем в массив rowspan.
func (r *Row) updateRowSpans(rowspan int) {
	r.rowspans[r.columnIndex] += rowspan
}

// При создании новой строки уменьшаем все rowspan на 1.
func (r *Row) decrementRowSpans() []int {
	for i := range r.rowspans {
		r.rowspans[i]--
	}

	return r.rowspans
}
