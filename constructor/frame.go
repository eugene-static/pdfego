package constructor

import (
	"slices"
)

type Frame struct {
	parent frameParent
	fields []*Field
}

func NewFrame() *Frame {
	return &Frame{}
}

func (f *Frame) Fields(fields ...*Field) *Frame {
	for i := range fields {
		fields[i].parent = f
	}

	f.fields = fields

	return f
}

func (f *Frame) setParentCell(parent *Field) {
	f.parent = parent
}

func (f *Frame) width() (w float64) {
	for i := range f.fields {
		w += f.fields[i].width()
	}

	return w
}

func (f *Frame) height() (h float64) {
	for i := range f.fields {
		hField := f.fields[i].height()
		if h < hField {
			h = hField
		}
	}

	return h
}

//func (f *Frame) Draw(core *fpdf.Fpdf) {
//	x, y := core.GetXY()
//
//	for i := range f.fields {
//		f.fields[i].Draw(core)
//	}
//
//	core.SetXY(x, y+f.height())
//}

func (f *Frame) cells() (cells []*TableCell) {
	for i := range f.fields {
		slices.Concat(cells, f.fields[i].cells())
	}

	return cells
}

type Field struct {
	parent      *Frame
	drawers     []Drawer
	unbreakable bool
}

func NewField() *Field {
	return &Field{}
}

func (f *Field) Drawers(drawers ...Drawer) *Field {
	f.drawers = drawers

	return f
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

//func (f *Field) Draw(core *fpdf.Fpdf) {
//	x, y := core.GetXY()
//
//	for i := range f.drawers {
//		f.drawers[i].Draw(core)
//	}
//
//	//pageW, _ := core.GetPageSize()
//	//marginL, _, marginR, _ := core.GetMargins()
//	//pageW -= marginL + marginR
//
//	core.SetXY(x+f.width(), y)
//}

func (f *Field) cells() (cells []*TableCell) {
	for i := range f.drawers {
		cells = slices.Concat(cells, f.drawers[i].cells())
	}

	return cells
}

func (f *Field) isParent() frameParent {
	return f
}
