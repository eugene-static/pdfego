package tempst

import "codeberg.org/go-pdf/fpdf"

type Tmpl struct {
	core    *fpdf.Fpdf
	width   float64
	height  float64
	margins float64
	grids   []*Grid
}

func (t *Tmpl) Grids(grids ...*Grid) {
	t.grids = grids
}

func (t *Tmpl) Draw() {
	for i := range t.grids {
		t.grids[i].draw(t.core)
	}
}

type Grid struct {
	cols  int64
	rows  int64
	cells []*GridCell
}

func NewGrid(cols, rows int64) *Grid {
	return &Grid{cols: cols, rows: rows}
}

func (g *Grid) Cells(cells ...*GridCell) *Grid {
	g.cells = cells

	return g
}

func (g *Grid) draw(core *fpdf.Fpdf) {
	for i := range g.cells {
		g.cells[i].grid = g

		g.cells[i].draw(core)
	}
}

type GridCell struct {
	grid        *Grid
	tables      []*Table
	x           float64
	y           float64
	unbreakable bool
}

func NewGridCell() *GridCell {
	return &GridCell{}
}

func (gc *GridCell) Tables(tables ...*Table) *GridCell {
	gc.tables = tables

	return gc
}

func (gc *GridCell) draw(core *fpdf.Fpdf) {
	for i := range gc.tables {
		gc.tables[i].gridCell = gc

		gc.tables[i].draw(core)
	}
}

func vals(values ...float64) []float64 {
	return values
}
