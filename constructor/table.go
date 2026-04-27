package constructor

import (
	"strings"

	"codeberg.org/go-pdf/fpdf"
)

type Table struct {
	gridCell *Field
	cols     []float64
	rows     []*Row
}

func NewTable(cols []float64, rows ...*Row) *Table {
	t := &Table{
		cols: cols,
		rows: rows,
	}

	return t
}

func (t *Table) SetParentCell(parent *Field) {
	t.gridCell = parent
}

func (t *Table) width() (w float64) {
	for _, col := range t.cols {
		w += col
	}

	return w
}

func (t *Table) height() (h float64) {
	for _, row := range t.rows {
		h += row.h
	}

	return h
}

func (t *Table) Draw(core *fpdf.Fpdf) {
	for i := range t.rows {
		t.rows[i].table = t

		t.rows[i].Draw(core)
	}
}

func (t *Table) cellWidth(index, colspan int) (w float64) {
	bound := index + colspan
	if bound > len(t.cols) {
		bound = len(t.cols)
	}

	for i := index; i < bound; i++ {
		w += t.cols[i]
	}

	return w
}

type Row struct {
	table *Table
	h     float64
	cells []*Cell
}

func NewRow(cells ...*Cell) *Row {
	return &Row{
		cells: cells,
	}
}

func (r *Row) Draw(core *fpdf.Fpdf) {
	x, y := core.GetXY()

	for i, cell := range r.cells {
		if cell.opts.Height > r.h {
			r.h = cell.opts.Height
		}

		cell.w = r.table.cellWidth(i, cell.opts.Colspan)

		cell.row = r

		cell.Draw(core)
	}

	core.SetXY(x, y+r.h)
}

type Cell struct {
	row  *Row
	w    float64
	text string
	opts CellOpts
}

func NewCell(text string, opts ...CellOpts) *Cell {
	tc := &Cell{
		text: text,
		opts: defaultCellOpts(),
	}

	if opts != nil {
		opt := opts[0]

		if opt.Height > 0 {
			tc.opts.Height = opt.Height
		}

		if opt.Colspan != 0 {
			tc.opts.Colspan = opt.Colspan
		}

		if opt.Rowspan != 0 {
			tc.opts.Rowspan = opt.Rowspan
		}

		if opt.Align != "" {
			tc.opts.Align = opt.Align
		}

		if opt.Border != "" {
			tc.opts.Border = opt.Border
		}

		if opt.FontSize != 0 {
			tc.opts.FontSize = opt.FontSize
		}

		if opt.Style != "" {
			tc.opts.Style = opt.Style
		}

		tc.opts.Wrap = opt.Wrap
	}

	return tc
}

func NewFormCell(text string, opts ...CellOpts) *Cell {
	if opts != nil {
		opts[0].Border = "B"
	}

	return NewCell(text, opts...)
}

func (c *Cell) Draw(core *fpdf.Fpdf) {
	if c.opts.FontSize > 0 {
		fontSize, _ := core.GetFontSize()

		core.SetFontSize(c.opts.FontSize)
		defer core.SetFontSize(fontSize)
	}

	if strings.ContainsRune(c.opts.Style, 'B') {
		core.SetFontStyle("B")
		defer core.SetFontStyle("")
	}

	if strings.ContainsRune(c.opts.Border, '+') {
		lineWidth := core.GetLineWidth()

		core.SetLineWidth(0.4)
		defer core.SetLineWidth(lineWidth)

		c.opts.Border = strings.ReplaceAll(c.opts.Border, "+", "")
	}

	_, fontSize := core.GetFontSize()
	lineHeight := c.lineHeight(fontSize)

	if c.opts.Height == 0 {
		c.opts.Height = c.row.h

		if c.row.h == 0 {
			c.opts.Height = lineHeight
		}
	}

	x, y := core.GetXY()

	if c.w == 0 {
		wPage, _ := core.GetPageSize()
		_, _, rMargin, _ := core.GetMargins()

		wPage -= rMargin

		c.w = wPage - x
	}

	if c.opts.Wrap {
		core.CellFormat(c.w, c.opts.Height, "", c.opts.Border, 0, c.opts.Align, false, 0, "")

		splitText := core.SplitText(c.text, c.w)

		vOffset := c.yOffset(len(splitText), lineHeight)

		core.SetXY(x, y+vOffset)

		for _, line := range splitText {
			x, y = core.GetXY()
			core.CellFormat(c.w, lineHeight, line, "", 0, c.opts.Align, false, 0, "")
			core.SetXY(x, y+lineHeight)
		}

		core.SetXY(x+c.w, y)

		return
	}

	core.CellFormat(c.w, c.opts.Height, c.text, c.opts.Border, 0, c.opts.Align, false, 0, "")
}

func (c *Cell) yOffset(linesNum int, lineHeight float64) float64 {
	vOffset := 0.0

	if strings.ContainsRune(c.opts.Align, 'M') {
		vOffset = (c.opts.Height - float64(linesNum)*lineHeight) / 2
	}

	if strings.ContainsRune(c.opts.Align, 'B') {
		vOffset = c.opts.Height - float64(linesNum)*lineHeight
	}

	return vOffset
}

func (c *Cell) lineHeight(fontSize float64) float64 {
	return fontSize * 1.2
}

type CellOpts struct {
	Height   float64
	Align    string
	Border   string
	Colspan  int
	Rowspan  int
	Style    string
	FontSize float64
	Wrap     bool
}

func defaultCellOpts() CellOpts {
	return CellOpts{
		//Height:  tc.lineHeight() ,
		Align:    "CM",
		Border:   "",
		Colspan:  1,
		Rowspan:  1,
		Style:    "",
		FontSize: 0,
		Wrap:     false,
	}
}
