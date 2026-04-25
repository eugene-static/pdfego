package tempst

import (
	"strings"

	"codeberg.org/go-pdf/fpdf"
)

type Table struct {
	gridCell *GridCell
	cols     []float64
	width    float64
	rows     []*Row
}

func NewTable(cols []float64) *Table {
	t := &Table{
		cols: cols,
	}

	t.setWidth()

	return t
}

func (t *Table) Rows(rows ...*Row) *Table {
	t.rows = rows

	return t
}

func (t *Table) draw(core *fpdf.Fpdf) {
	for i := range t.rows {
		t.rows[i].table = t

		t.rows[i].draw(core)
	}
}

func (t *Table) setWidth() {
	for _, col := range t.cols {
		t.width += col
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
	cells []*TableCell
}

func NewRow() *Row {
	return &Row{}
}

func (r *Row) Cells(cells ...*TableCell) *Row {
	r.cells = cells

	return r
}

func (r *Row) draw(core *fpdf.Fpdf) {
	for i, cell := range r.cells {
		if cell.h > r.h {
			r.h = cell.h
		}

		cell.w = r.table.cellWidth(i, cell.colspan)

		cell.row = r

		cell.draw(core)
	}

	core.Ln(r.h)
}

type TableCell struct {
	row     *Row
	w       float64
	h       float64
	align   string
	border  string
	colspan int
	content Text
}

func NewTableCell(opts ...TableCellOpts) *TableCell {
	tc := &TableCell{
		colspan: 1,
	}

	if opts != nil {
		opt := opts[0]

		if opt.Height > 0 {
			tc.h = opt.Height
		}

		tc.align = opt.Align
		tc.border = opt.Border

		if opt.Colspan > 0 {
			tc.colspan = opt.Colspan
		}
	}

	return tc
}

func (tc *TableCell) SetString(s string) *TableCell {
	tc.content = Text{
		body: s,
		wrap: false,
	}

	return tc
}

func (tc *TableCell) SetText(text Text) *TableCell {
	tc.content = text

	return tc
}

func (tc *TableCell) Align(align string) *TableCell {
	tc.align = align

	return tc
}

func (tc *TableCell) draw(core *fpdf.Fpdf) {
	if tc.content.fontSize > 0 {
		fontSize, _ := core.GetFontSize()

		core.SetFontSize(tc.content.fontSize)
		defer core.SetFontSize(fontSize)
	}

	if strings.ContainsRune(tc.content.style, 'B') {
		core.SetFontStyle("B")
		defer core.SetFontStyle("")
	}

	if strings.ContainsRune(tc.border, '+') {
		lineWidth := core.GetLineWidth()

		core.SetLineWidth(0.4)
		defer core.SetLineWidth(lineWidth)

		tc.border = strings.ReplaceAll(tc.border, "+", "")
	}

	_, fontSize := core.GetFontSize()
	lineHeight := tc.lineHeight(fontSize)

	if tc.h == 0 {
		tc.h = tc.row.h

		if tc.row.h == 0 {
			tc.h = lineHeight
		}
	}

	if tc.content.wrap {
		x, y := core.GetXY()

		core.CellFormat(tc.w, tc.h, "", tc.border, 0, tc.align, false, 0, "")

		splitText := core.SplitText(tc.content.body, tc.w)

		vOffset := tc.vOffset(len(splitText), lineHeight)

		core.SetXY(x, y+vOffset)

		for _, line := range splitText {
			core.CellFormat(tc.w, lineHeight, line, "", 0, tc.align, false, 0, "")
			core.Ln(-1)
		}

		core.SetXY(x+tc.w, y)

		return
	}

	core.CellFormat(tc.w, tc.h, tc.content.body, tc.border, 0, tc.align, false, 0, "")

	//x := core.GetX()
	//
	//if x >= tc.row.table.width {
	//	core.Ln(tc.h)
	//}
}

func (tc *TableCell) vOffset(linesNum int, lineHeight float64) float64 {
	vOffset := 0.0

	if strings.ContainsRune(tc.align, 'M') {
		vOffset = (tc.h - float64(linesNum)*lineHeight) / 2
	}

	if strings.ContainsRune(tc.align, 'B') {
		vOffset = tc.h - float64(linesNum)*lineHeight
	}

	return vOffset
}

func (tc *TableCell) lineHeight(fontSize float64) float64 {
	return fontSize * 1.2
}

type TableCellOpts struct {
	Height  float64
	Align   string
	Border  string
	Colspan int
	Rowspan int
}
