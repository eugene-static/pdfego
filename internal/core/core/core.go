package core

import (
	"compress/zlib"
	"log/slog"

	"github.com/eugene-static/pdf-craft/internal/buffer"
	"github.com/eugene-static/pdf-craft/internal/font"
	"github.com/eugene-static/pdf-craft/pkg/unit"
)

const (
	FontRegular = "REG"
	FontBold    = "BOLD"
	FontItalic  = "ITALIC"

	Portrait  = "P"
	Landscape = "L"
)

const (
	bufferSize = 1 << 16
)

type Core struct {
	log        *slog.Logger
	mainBuffer *buffer.Buffer
	//headBuffer *buffer.Buffer
	//pageBuffer *buffer.Buffer
	compressor compressor
	cursor     cursor
	page       Page
	border     border
	fonts      map[string]*font.Font
	fontSize   unit.PT
	//compress   bool
	pagesCount int64
	cellCount  int
	offsets    []int
	pageObjs   []int64
	error      error
}

type cursor struct {
	x, y float64
}

type border struct {
	thin  unit.PT
	thick unit.PT
}

type compressor struct {
	buffer *buffer.Buffer
	writer *zlib.Writer
}

func New(orientation string) *Core {
	pg := Page{
		width:  unit.PT(595.2).MM(),
		height: unit.PT(841.89).MM(),
		buffer: buffer.New(bufferSize),
	}

	if orientation == Landscape {
		pg.width, pg.height = pg.height, pg.width
	}

	offsets := make([]int, 3, 100)

	compBuffer := buffer.New(bufferSize)
	comp := compressor{
		buffer: compBuffer,
		writer: zlib.NewWriter(compBuffer),
	}

	return &Core{
		mainBuffer: buffer.New(bufferSize),
		//pageBuffer: buffer.New(1 << 16),
		compressor: comp,
		fonts:      make(map[string]*font.Font),
		page:       pg,
		offsets:    offsets,
	}
}

func (core *Core) WriteError(err error) {
	core.error = err
}

func (core *Core) Err() error {
	return core.error
}

func (core *Core) SetBorders(thin, thick float64) {
	core.border = border{
		thin:  unit.PT(thin),
		thick: unit.PT(thick),
	}
}

func (core *Core) SetLogger(log *slog.Logger) {
	core.log = log
}

//func (core *Core) Compress() {
//	core.compress = true
//}

func (core *Core) SetFont(path, alias string) error {
	f, err := font.New(path, alias)
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
	core.fontSize = unit.PT(fontSize)
}

func (core *Core) IncreaseCellsCount(n int) {
	core.cellCount += n
}

func (core *Core) DefaultFontSize() unit.PT {
	return core.fontSize
}

func (core *Core) BorderThin() unit.PT {
	return core.border.thin
}

func (core *Core) BorderThick() unit.PT {
	return core.border.thick
}

func (core *Core) Bytes() []byte {
	defer core.mainBuffer.Reset()
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

func (comp compressor) Compress(buf []byte) ([]byte, error) {
	comp.buffer.Reset()
	comp.writer.Reset(comp.buffer)

	_, err := comp.writer.Write(buf)
	if err != nil {
		return nil, err
	}

	err = comp.writer.Close()
	if err != nil {
		return nil, err
	}

	return comp.buffer.Bytes(), nil
}
