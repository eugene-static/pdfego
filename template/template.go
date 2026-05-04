package template

type Template struct {
	core   *Core
	frames []Frame
}

func (t *Template) Frame() *Frame {
	frame := Frame{}

	t.frames = append(t.frames, frame)

	return &frame
}

type Drawer interface {
	Draw()
}

type Frame struct {
	core   *Core
	fields []Field
}

func (f *Frame) Field() *Field {
	field := Field{
		core: f.core,
	}

	f.fields = append(f.fields, field)

	return &field
}

func (f *Frame) Draw() {
	for _, field := range f.fields {
		for _, drawer := range field.drawers {
			drawer.Draw()
		}
	}
}

type Field struct {
	core    *Core
	drawers []Drawer
}

func (f *Field) Frame() *Frame {
	frame := Frame{
		core: f.core,
	}

	f.drawers = append(f.drawers, &frame)

	return &frame
}

func (f *Field) Table(cols ...float64) *Table {
	table := Table{
		core: f.core,
		cols: cols,
		//rowspans: make([]int, len(cols)),
	}

	f.drawers = append(f.drawers, &table)

	return &table
}

type Table struct {
	core     *Core
	x        float64
	y        float64
	cols     []float64
	rowIndex int
	rows     []Row
	//rowspans []int
}

func (t *Table) Row() Row {
	r := Row{
		core:     t.core,
		x:        t.x,
		cols:     t.cols,
		height:   t.core.fontHeight,
		cells:    make([]cell, len(t.cols)),
		rowspans: make([]int, len(t.cols)),
	}

	if t.rowIndex > 0 {
		r.rowspans = t.rows[t.rowIndex-1].decrementRowspans()
	}

	t.rowIndex++

	return r
}

func (t *Table) Draw() {
	height := 0.0
	x := t.x

	for _, row := range t.rows {
		row.y = t.y + height

		for _, c := range row.cells {
			c.x = x
			x += c.width

			if !c.busy {
				continue
			}

			c.height = t.cellHeight(c.rowspan)
			height += row.height

			c.render()
		}
	}
}

func (t *Table) cellHeight(rowspan int) (h float64) {
	for i := t.rowIndex; i < len(t.rows) || i < t.rowIndex+rowspan; i++ {
		h += t.rows[i].height
	}

	return h
}

// [3, 1, 1, 2, 1]
// [2, 0, 0, 1, 0] -> [2, 1, 1, 1, 2]
// [1, 0, 0, 0, 1] -> [1, 1, 1, 1, 1]
// [0...........0] -> [1...........1]
