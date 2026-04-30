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
	}

	f.drawers = append(f.drawers, &table)

	return &table
}

type Table struct {
	core *Core
	cols []float64
	rows []Row
}

func (t *Table) Row() Row {
	return Row{
		core:   t.core,
		cols:   t.cols,
		height: t.core.fontHeight,
	}
}

func (t *Table) Draw() {
	//cells := make([]cell, 0, len(t.cols)*len(t.rows))

	for _, row := range t.rows {
		if len(row.cells) > len(t.cols) {
			//TODO: write warn
			break
		}

		for i := range row.cells {
			row.cells[i].height = row.height

			row.cells[i].draw()
		}
	}
}
