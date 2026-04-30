package template

import (
	"bytes"
	"fmt"
	"os"

	"github.com/golang/freetype/truetype"
	fontface "golang.org/x/image/font"
)

const (
	FontRegular = "R"
	FontBold    = "B"
	FontItalic  = "I"

	Portrait  = "P"
	Landscape = "L"
)

type Core struct {
	buffer             *bytes.Buffer
	pagesContentStream *bytes.Buffer
	cursor             cursor
	page               page
	border             border
	fonts              map[string]*font
	fontSize           float64
	fontHeight         float64
	pagesCount         int64
	offsets            []int
}

type cursor struct {
	x, y float64
}

type page struct {
	width, height float64
	margin        float64
}

type font struct {
	alias string
	font  *truetype.Font
	face  fontface.Face
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

	return &Core{
		buffer:  new(bytes.Buffer),
		page:    pg,
		offsets: make([]int, 3, 100),
	}
}

func (core *Core) SetMargin(margin float64) {
	core.page.margin = margin
}

func (core *Core) SetDefaultFontSize(fontSize float64) {
	core.fontSize = fontSize
	core.fontHeight = mm(fontSize) * 1.2
}

func (core *Core) SetFont(path, alias string) error {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fontParsed, err := truetype.Parse(fontBytes)
	if err != nil {
		return err
	}

	core.fonts = make(map[string]*font)

	core.fonts[alias] = &font{
		alias: alias,
		font:  fontParsed,
	}

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

func (core *Core) SetBorders(thin, thick float64) {
	core.border = border{
		thin:  thin,
		thick: thick,
	}
}

func (core *Core) setFontFace(alias string, face fontface.Face) {
	core.fonts[alias].face = face
}

func (core *Core) x() float64 {
	return core.cursor.x
}

func (core *Core) y() float64 {
	return core.page.height - core.cursor.y
}

func (core *Core) xy() (float64, float64) {
	return core.x(), core.y()
}

//type printer interface {
//	print(s ...string)
//	printf(s string, args ...any)
//	printInt64(v int64)
//	printFloat64(v float64)
//	printText(fontAlias, text string)
//	printSpace()
//	printFont(font string, fontSize float64)
//	printXY(x, y float64)
//	printLine(bw, x0, y0, x1, y1 float64)
//}

// TODO: подумать, как написать эти же функции, но для буфера страницы. Это нужно для того, чтобы в начале stream писать его длину <</Length %d>>
func (core *Core) print(s ...string) {
	for i := range s {
		core.buffer.WriteString(s[i])
	}
}

func (core *Core) println(s ...string) {
	for i := range s {
		core.buffer.WriteString(s[i])
		core.buffer.WriteByte('\n')
	}
}

func (core *Core) printf(s string, args ...any) {
	fmt.Fprintf(core.buffer, s, args...)
}

func (core *Core) printText(fontAlias, text string) {
	core.buffer.WriteByte('<')
	defer core.buffer.WriteByte('>')

	for _, r := range text {
		idx := core.getFont(fontAlias).font.Index(r)

		appendHex4(core.buffer, uint16(idx))
	}
}

func (core *Core) printFloat64(v float64) {
	appendFloat(core.buffer, pt(v))
}

func (core *Core) printInt64(v int64) {
	appendInt(core.buffer, v)
}

func (core *Core) printSpace() {
	core.buffer.WriteByte(' ')
}

func (core *Core) printFont(font string, fontSize float64) {
	core.print("/", font, " ")
	core.printFloat64(fontSize)
	core.print(" Tf ")
}

func (core *Core) printXY(x, y float64) {
	core.printFloat64(x)
	core.printSpace()
	core.printFloat64(y)
}

func (core *Core) printLine(bw, x0, y0, x1, y1 float64) {
	// 2 w x0 y0 m x1 y1 l S
	core.printFloat64(bw)
	core.print(" w ")
	core.printXY(x0, y0)
	core.print(" m ")
	core.printXY(x1, y1)
	core.print(" l S ")
}

func (core *Core) measureText(fontAlias string, fontSize float64, text string) float64 {
	f := core.getFont(fontAlias)

	if f.face == nil {
		f.face = truetype.NewFace(f.font, &truetype.Options{
			Size:    fontSize,
			DPI:     dpi,
			Hinting: fontface.HintingFull,
		})

		core.setFontFace(fontAlias, f.face)
	}

	width := float64(fontface.MeasureString(f.face, text)) / mm(64)

	return width
}

func (core *Core) getFont(alias string) *font {
	return core.fonts[alias]
}

func (core *Core) fontRegular() *font {
	return core.fonts["R"]
}

func (core *Core) fontBold() *font {
	return core.fonts["B"]
}

func (core *Core) appendOffset() {
	xLen := core.buffer.Len()

	core.offsets = append(core.offsets, xLen)
}

func (core *Core) getObjNum() int64 {
	return int64(len(core.offsets))
}
