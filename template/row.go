package template

import "strings"

type Row struct {
	core    *Core
	height  float64
	cols    []float64
	cellIdx int
	cells   []cell
}

func (r *Row) Cell(text string, opts ...CellOpts) {
	opt := CellOpts{}
	if opts != nil {
		opt = opts[0]
	}

	c := r.cell(text, opt)

	r.cellIdx++
	if r.cellIdx >= len(r.cols) {
		return
	}

	r.cells = append(r.cells, c)
}

func (r *Row) cell(text string, opts CellOpts) cell {
	c := cell{
		core:       r.core,
		width:      r.cols[r.cellIdx],
		font:       FontRegular,
		fontSize:   r.core.fontSize,
		border:     "",
		borderSize: r.core.border.thin,
		align:      "CM",
	}

	if c.width == -1 {
		c.width = c.core.page.width - c.core.page.margin - c.x
	}

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
	// Поэтому высота ячейки будет определяться в методе draw().

	c.text = cellText

	return c
}
