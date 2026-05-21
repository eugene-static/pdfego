package core

import (
	"log/slog"

	"github.com/eugene-static/pdf-craft/buffer"
	"github.com/eugene-static/pdf-craft/font"
	"github.com/eugene-static/pdf-craft/meter"
)

const (
	FontRegular = "REG"
	FontBold    = "BOLD"
	FontItalic  = "ITALIC"

	Portrait  = "P"
	Landscape = "L"
)

type Core struct {
	log        *slog.Logger
	mainBuffer *buffer.Buffer
	headBuffer *buffer.Buffer
	pageBuffer *buffer.Buffer
	cursor     cursor
	page       Page
	border     border
	fonts      map[string]*font.Font
	fontSize   meter.PT
	compress   bool
	pagesCount int64
	cellCount  int
	offsets    []int
	pageObjs   []int64
}

type cursor struct {
	x, y float64
}

type border struct {
	thin  meter.PT
	thick meter.PT
}

func New(orientation string) *Core {
	pg := Page{
		width:  meter.PT(595.2).MM(),
		height: meter.PT(841.89).MM(),
	}

	if orientation == Landscape {
		pg.width, pg.height = pg.height, pg.width
	}

	offsets := make([]int, 3, 100)

	return &Core{
		mainBuffer: buffer.New(),
		pageBuffer: buffer.New(),
		fonts:      make(map[string]*font.Font),
		page:       pg,
		offsets:    offsets,
	}
}

func (core *Core) SetBorders(thin, thick float64) {
	core.border = border{
		thin:  meter.PT(thin),
		thick: meter.PT(thick),
	}
}

func (core *Core) SetLogger(log *slog.Logger) {
	core.log = log
}

func (core *Core) Compress() {
	core.compress = true
}

func (core *Core) SetFont(path, alias string) error {
	f, err := font.NewFont(path, alias)
	if err != nil {
		return err
	}

	core.fonts[alias] = f

	return nil
}

func (core *Core) SetFontRegular(path string) error {
	err := core.SetFont(path, FontRegular)
	if err != nil {
		return err
	}

	return nil
}

func (core *Core) SetFontBold(path string) error {
	err := core.SetFont(path, FontBold)
	if err != nil {
		return err
	}

	return nil
}

func (core *Core) SetDefaultFontSize(fontSize int) {
	core.fontSize = meter.PT(fontSize)
}

func (core *Core) IncreaseCellsCount(n int) {
	core.cellCount += n
}

func (core *Core) DefaultFontSize() meter.PT {
	return core.fontSize
}

func (core *Core) BorderThin() meter.PT {
	return core.border.thin
}

func (core *Core) BorderThick() meter.PT {
	return core.border.thick
}

func (core *Core) Bytes() []byte {
	return core.mainBuffer.Bytes()
}

func (core *Core) Font(alias string) *font.Font {

	return core.fonts[alias]
}

func (core *Core) Log() *slog.Logger {
	return core.log
}

func (core *Core) newObject() int64 {
	objNum := int64(len(core.offsets))

	xLen := core.mainBuffer.Len()

	core.offsets = append(core.offsets, xLen)

	return objNum
}

func (core *Core) setObject(objNum int) {
	xLen := core.mainBuffer.Len()

	core.offsets[objNum] = xLen
}

func (core *Core) fontRegular() *font.Font {
	return core.fonts[FontRegular]
}

func (core *Core) fontBold() *font.Font {
	return core.fonts[FontBold]
}
