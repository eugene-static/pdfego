package constructor

import (
	"slices"
	"strings"

	"codeberg.org/go-pdf/fpdf"
)

type Table struct {
	field *Field
	cols  []float64
	rows  []*Row
}

func NewTable(cols ...float64) *Table {
	t := &Table{
		cols: cols,
	}

	return t
}

func (t *Table) Rows(rows ...*Row) *Table {
	for i := range rows {
		rows[i].table = t
	}

	t.rows = rows

	return t
}

//func (t *Table) Draw(core *fpdf.Fpdf) {
//	for i := range t.rows {
//		t.rows[i].Draw(core)
//	}
//}

func (t *Table) setParentCell(parent *Field) {
	t.field = parent
}

func (t *Table) width() (w float64) {
	for _, col := range t.cols {
		w += col
	}

	return w
}

func (t *Table) height() (h float64) {
	for _, row := range t.rows {
		h += row.height
	}

	return h
}

func (t *Table) cells() (cells []*TableCell) {
	for i := range t.rows {
		slices.Concat(cells, t.rows[i].cells)
	}

	return cells
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
	table      *Table
	lineHeight float64
	height     float64
	cells      []*TableCell
}

func NewRow() *Row {
	r := &Row{}
	return r
}

func (r *Row) Cells(cells ...*TableCell) *Row {
	for i := range cells {
		maxHeight := max(cells[i].opts.Height, r.height)

		cells[i].opts.Height, r.height = maxHeight, maxHeight

		cells[i].row = r
	}

	r.cells = cells

	return r
}

//func (r *Row) Draw(core *fpdf.Fpdf) {
//	x, y := core.GetXY()
//	_, fontSize := core.GetFontSize()
//
//	r.lineHeight = lineHeight(fontSize)
//
//	for i := range r.cells {
//		r.cells[i].w = r.table.cellWidth(i, r.cells[i].opts.Colspan)
//
//		if r.cells[i].opts.Wrap {
//			wrappedHeight := float64(len(core.SplitText(r.cells[i].text, r.cells[i].w))) * r.lineHeight
//
//			if wrappedHeight > r.height {
//				r.height = wrappedHeight
//			}
//		}
//
//		r.cells[i].Draw(core)
//	}
//
//	core.SetXY(x, y+r.height)
//}

type TableCell struct {
	row  *Row
	x    float64
	y    float64
	w    float64
	h    float64
	text string
	opts CellOpts
}

func Cell(text string, opts ...CellOpts) *TableCell {
	tc := &TableCell{
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

func Form(text string, opts ...CellOpts) *TableCell {
	if opts != nil {
		opts[0].Border = "B"
	}

	return Cell(text, opts...)
}

func Text(text string, opts ...CellOpts) *TableCell {
	if opts != nil {
		opts[0].Border = ""
	}

	return Cell(text, opts...)
}

func (c *TableCell) Draw(core *fpdf.Fpdf) {
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

	x, y := core.GetXY()
	c.x, c.y = x, y

	if c.w == 0 {
		wPage, _ := core.GetPageSize()
		_, _, rMargin, _ := core.GetMargins()

		wPage -= rMargin

		c.w = wPage - x
	}

	if c.opts.Wrap {
		core.CellFormat(c.w, c.opts.Height, "", c.opts.Border, 0, c.opts.Align, false, 0, "")

		splitText := core.SplitText(c.text, c.w)

		vOffset := c.yOffset(len(splitText), c.row.lineHeight)

		core.SetXY(x, y+vOffset)

		for _, line := range splitText {
			x, y = core.GetXY()
			core.CellFormat(c.w, c.row.lineHeight, line, "", 0, c.opts.Align, false, 0, "")
			core.SetXY(x, y+c.row.lineHeight)
		}

		core.SetXY(x+c.w, y)

		return
	}

	core.CellFormat(c.w, c.opts.Height, c.text, c.opts.Border, 0, c.opts.Align, false, 0, "")
}

func (c *TableCell) yOffset(linesNum int, lineHeight float64) float64 {
	vOffset := 0.0

	if strings.ContainsRune(c.opts.Align, 'M') {
		vOffset = (c.opts.Height - float64(linesNum)*lineHeight) / 2
	}

	if strings.ContainsRune(c.opts.Align, 'B') {
		vOffset = c.opts.Height - float64(linesNum)*lineHeight
	}

	return vOffset
}

func lineHeight(fontSize float64) float64 {
	return fontSize * 1.2
}

type CellOpts struct {
	Height      float64
	Colspan     int
	Rowspan     int
	Align       string
	Border      string
	BorderWidth float64
	Style       string
	FontSize    float64
	Wrap        bool
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
