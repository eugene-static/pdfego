package constructor

import "codeberg.org/go-pdf/fpdf"

type Grid struct {
	parent GridParent
	fields []*Field
}

func NewGrid(cells ...*Field) *Grid {
	return &Grid{
		fields: cells}
}

func (g *Grid) SetParentCell(parent *Field) {
	g.parent = parent
}

func (g *Grid) width() (w float64) {
	for i := range g.fields {
		w += g.fields[i].width()
	}

	return w
}

func (g *Grid) height() (h float64) {
	for i := range g.fields {
		hGidCell := g.fields[i].height()
		if h < hGidCell {
			h = hGidCell
		}
	}

	return h
}

func (g *Grid) Draw(core *fpdf.Fpdf) {
	x, y := core.GetXY()
	//g.x, g.y = x, y
	//hGrid := 0.0

	for i := range g.fields {
		g.fields[i].Draw(core)

		//hGidCell := g.fields[i].height()
		//if hGrid < hGidCell {
		//	hGrid = hGidCell
		//}
	}

	core.SetXY(x, y+g.height())
}

type Field struct {
	parent      *Grid
	drawers     []Drawer
	unbreakable bool
}

func NewField(drawers ...Drawer) *Field {
	gc := &Field{
		drawers: drawers,
	}

	//for i := range gc.drawers {
	//	gc.drawers[i].SetParentCell(gc)
	//}

	return gc
}

func (f *Field) width() (w float64) {
	for _, drawer := range f.drawers {
		w += drawer.width()
	}

	return w
}

func (f *Field) height() (h float64) {
	for _, drawer := range f.drawers {
		h += drawer.height()
	}

	return h
}

func (f *Field) Draw(core *fpdf.Fpdf) {
	x, y := core.GetXY()

	for i := range f.drawers {
		f.drawers[i].SetParentCell(f)
		f.drawers[i].Draw(core)
	}

	//pageW, _ := core.GetPageSize()
	//marginL, _, marginR, _ := core.GetMargins()
	//pageW -= marginL + marginR

	core.SetXY(x+f.width(), y)
}

func (f *Field) Parent() GridParent {
	return f
}
