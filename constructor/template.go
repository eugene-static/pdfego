package constructor

import "codeberg.org/go-pdf/fpdf"

type Template struct {
	core    *fpdf.Fpdf
	width   float64
	height  float64
	margins float64
	grids   []*Grid
}

func (t *Template) Grids(grids ...*Grid) {
	t.grids = grids
}

func (t *Template) Draw() {
	for i := range t.grids {
		t.grids[i].Draw(t.core)
	}
}

func (t *Template) Parent() GridParent {
	return t
}

type Drawer interface {
	Draw(core *fpdf.Fpdf)
	width() float64
	height() float64
	SetParentCell(parent *Field)
}

type GridParent interface {
	Parent() GridParent
}
