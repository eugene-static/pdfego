package template

import "log/slog"

const (
	FontRegular = "REG"
	FontBold    = "BOLD"
	FontItalic  = "ITALIC"

	Portrait  = "P"
	Landscape = "L"
)

type Core struct {
	log         *slog.Logger
	mainBuffer  *buffer
	headBuffer  *buffer
	pageBuffers []*buffer
	cursor      cursor
	page        page
	border      border
	fonts       map[string]*font
	fontSize    int
	fontHeight  float64
	pagesCount  int64
	offsets     []int
}

type cursor struct {
	x, y float64
}

type border struct {
	thin  float64
	thick float64
}

func New(orientation string) *Core {
	pg := page{
		width:  595.28,
		height: 841.89,
	}

	if orientation == Landscape {
		pg.width, pg.height = pg.height, pg.width
	}

	offsets := make([]int, 3, 100)

	return &Core{
		mainBuffer: newBuffer(),
		headBuffer: newBuffer(),
		fonts:      make(map[string]*font),
		page:       pg,
		offsets:    offsets,
	}
}

func (core *Core) SetBorders(thin, thick float64) {
	core.border = border{
		thin:  thin,
		thick: thick,
	}
}

func (core *Core) SetLogger(log *slog.Logger) {
	core.log = log
}

func (core *Core) x() float64 {
	return core.cursor.x
}

func (core *Core) y() float64 {
	return core.cursor.y
}

func (core *Core) x0() float64 {
	return core.page.margin
}

func (core *Core) y0() float64 {
	return core.page.margin
}

func (core *Core) xy() (float64, float64) {
	return core.x(), core.y()
}

func (core *Core) x0y0() (float64, float64) {
	return core.x0(), core.y0()
}

func (core *Core) setX(x float64) {
	core.cursor.x = x
}

func (core *Core) setY(y float64) {
	core.cursor.y = y
}

func (core *Core) setXY(x, y float64) {
	core.setX(x)
	core.setY(y)
}

func (core *Core) appendOffset() {
	xLen := core.mainBuffer.content.Len()

	core.offsets = append(core.offsets, xLen)
}

func (core *Core) setOffset(offset int) {
	xLen := core.mainBuffer.content.Len()

	core.offsets[offset] = xLen
}

func (core *Core) getObjNum() int64 {
	return int64(len(core.offsets))
}

func (core *Core) ptX(x float64) float64 {
	return pt(x)
}

func (core *Core) ptY(y float64) float64 {
	return core.page.height - pt(y)
}

func (core *Core) ptXY(x, y float64) (float64, float64) {
	return core.ptX(x), core.ptY(y)
}
