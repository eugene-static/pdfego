package template

import "strings"

type Row struct {
	core      *Core
	x         float64
	y         float64
	height    float64
	cols      []float64
	cellIndex int
	cells     []cell
	rowspans  []int
}

func (r *Row) Cell(text string, opts ...CellOpts) {
	for i := r.cellIndex; i < len(r.rowspans); i++ {
		if r.rowspans[i] > 0 {
			r.cellIndex++
		}
	}

	if r.cellIndex >= len(r.cols) {
		//TODO: сообщить или нет
		return
	}

	var opt CellOpts

	if opts != nil {
		opt = opts[0]
	}

	c := r.cell(text, opt)

	r.cells[r.cellIndex] = c

	r.rowspans[r.cellIndex] += c.rowspan

	for i := r.cellIndex; i < r.cellIndex+c.colspan; i++ {
		r.rowspans[i] += c.colspan
	}
	//r.cellIndex += c.colspan
}

func (r *Row) cell(text string, opts CellOpts) cell {
	c := cell{
		core:       r.core,
		font:       FontRegular,
		fontSize:   r.core.fontSize,
		border:     "",
		borderSize: r.core.border.thin,
		align:      "CM",
		colspan:    1,
		rowspan:    1,
		busy:       true,
	}

	//if c.width == -1 {
	//	c.width = c.core.page.width - c.core.page.margin - c.x
	//}

	if r.height < opts.Height {
		r.height = opts.Height
	}

	if opts.Font != "" {
		c.font = strings.ToUpper(opts.Font)
	}

	if opts.FontSize > 0 {
		c.fontSize = float64(opts.FontSize)
	}

	if opts.Border != "" {
		c.border = strings.ToUpper(opts.Border)
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

	c.width = r.cellWidth(c.colspan)

	cellText := []string{text}

	if opts.Wrap {
		split := c.core.splitText(c.font, text, c.fontSize, c.width)
		height := c.core.fontHeight * float64(len(split))

		if r.height < height {
			r.height = height
		}

		cellText = split
	}

	// Поле cell.height должно быть равно высоте строки, но пока все ячейки не будут созданы, мы не знаем итоговую высоту строки.
	// Поэтому высота ячейки будет определяться в методе render().

	c.text = cellText

	return c
}

func (r *Row) cellWidth(colspan int) (w float64) {
	for i := r.cellIndex; i < len(r.cols) || i < r.cellIndex+colspan; i++ {
		w += r.cols[i]
	}

	return w
}

func (r *Row) decrementRowspans() []int {
	for i := range r.rowspans {
		r.rowspans[i]--
	}

	return r.rowspans
}
