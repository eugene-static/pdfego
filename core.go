package pdf_craft

import (
	"compress/zlib"
	"fmt"
	"os"

	"github.com/eugene-static/pdf-craft/internal/buffer"
	"github.com/eugene-static/pdf-craft/internal/font"
	"github.com/eugene-static/pdf-craft/internal/image"
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
	mainBuffer *buffer.Buffer
	fonts      map[string]*font.Font
	images     map[string]*image.Image
	comp       compressor
	page       page
	fontSize   unit.PT
	borderSize unit.PT
	pagesCount int64
	offsets    []int
	pageObjs   []int64
	compress   bool
	error      error
}

func NewCore(orientation string) *Core {
	pg := page{
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
		comp:       comp,
		fonts:      make(map[string]*font.Font),
		images:     make(map[string]*image.Image),
		page:       pg,
		offsets:    offsets,
	}
}

func (core *Core) Compress() {
	core.compress = true
}

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

func (core *Core) SetDefaultBorderSize(size unit.PT) {
	core.borderSize = size
}

func (core *Core) SetMargin(margin unit.MM) {
	core.page.margin = margin
}

func (core *Core) DefaultFontSize() unit.PT {
	return core.fontSize
}

func (core *Core) DefaultBorderSize() unit.PT {
	return core.borderSize
}

func (core *Core) ReadImage(path, alias string) error {
	imageBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	img, err := image.New(alias, imageBytes)
	if err != nil {
		return err
	}

	core.images[alias] = img

	return nil
}

func (core *Core) AddImage(data []byte, alias string) error {
	img, err := image.New(alias, data)
	if err != nil {
		return err
	}

	core.images[alias] = img

	return nil
}

func (core *Core) writeError(err error) {
	core.error = err
}

func (core *Core) err() error {
	return core.error
}

func (core *Core) bytes() []byte {
	defer core.mainBuffer.Reset()
	return core.mainBuffer.Bytes()
}

func (core *Core) font(alias string) *font.Font {
	fnt, ok := core.fonts[alias]
	if !ok {
		err := fmt.Errorf("не найден шрифт с таким именем: %s", alias)

		core.writeError(err)

		return nil
	}

	return fnt
}

func (core *Core) image(alias string) *image.Image {
	img, ok := core.images[alias]
	if !ok {
		err := fmt.Errorf("не найдено изображение с таким именем: %s", alias)

		core.writeError(err)

		return nil
	}

	return img
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

func (core *Core) newPage() {
	if core.err() != nil {
		return
	}

	if core.page.headerBuffer != nil && core.page.headerBuffer.Len() > 0 {
		_, err := core.page.buffer.Write(core.page.headerBuffer.Bytes())
		if err != nil {
			core.writeError(err)

			return
		}
	}

	core.pagesCount++
}

func (core *Core) renderPage() {
	if core.page.buffer.Len() > 0 {
		core.writePage()

		core.page.buffer.Reset()
	}
}

type page struct {
	buffer       *buffer.Buffer
	headerBuffer *buffer.Buffer
	footerBuffer *buffer.Buffer
	width        unit.MM
	height       unit.MM
	margin       unit.MM
	marginLeft   unit.MM
	marginRight  unit.MM
	marginTop    unit.MM
	marginBottom unit.MM //TODO:
}

func (p *page) x0y0() (unit.MM, unit.MM) {
	return p.margin, p.margin - p.height
}

func (p *page) isBelowBottomBorder(y unit.MM) bool {
	return y+p.margin > 0
}

func (p *page) newHeader() *buffer.Buffer {
	if p.headerBuffer != nil {
		p.headerBuffer.Reset()

		return p.headerBuffer
	}

	buf := buffer.New(bufferSize)

	p.headerBuffer = buf

	return buf
}

func (p *page) removeHeader() {
	p.headerBuffer.Reset()
}

type compressor struct {
	buffer *buffer.Buffer
	writer *zlib.Writer
}

func (comp compressor) compress(buf []byte) ([]byte, error) {
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
