package constructor

import "codeberg.org/go-pdf/fpdf"

type Template struct {
	core    *fpdf.Fpdf
	width   float64
	height  float64
	margins float64
	frames  []*Frame
}

func (t *Template) Frames(frames ...*Frame) {
	t.frames = frames
}

func (t *Template) Draw() {
	cells := make([]*TableCell, 0)

	for i := range t.frames {
		cells = append(cells, t.frames[i].cells()...)
		//t.frames[i].Draw(t.core)
	}
}

func (t *Template) isParent() frameParent {
	return t
}

type Drawer interface {
	cells() []*TableCell
	width() float64
	height() float64
	setParentCell(parent *Field)
}

type frameParent interface {
	isParent() frameParent
}
