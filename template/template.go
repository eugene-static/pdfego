package template

import "codeberg.org/go-pdf/fpdf"

type Template struct {
	core *fpdf.Fpdf
}

type Table struct {
	cols []float64
	rows []Row
}

func (t *Table) Draw() {
	//cells := make([]Cell, 0, len(t.cols)*len(t.rows))

	for _, row := range t.rows {
		if len(row.cells) > len(t.cols) {
			//TODO: write warn
			break
		}

		for i := range row.cells {
			row.cells[i].width = t.cols[i]

			if row.cells[i].height < row.cells[i].opts.Height {
				row.cells[i].height = row.cells[i].opts.Height
			}

			if row.cells[i].opts.Wrap {
				//splitText := _
			}
		}
	}
}

type Row struct {
	height float64
	cells  []Cell
}

func (r *Row) Cell(text string, opts CellOpts) {
	cell := Cell{
		text: text,
		opts: opts,
	}

	r.cells = append(r.cells, cell)
}

type Cell struct {
	height float64
	width  float64
	text   string
	opts   CellOpts
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
