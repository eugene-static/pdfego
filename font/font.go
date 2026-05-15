package font

import (
	"bytes"
	"cmp"
	"compress/zlib"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/eugene-static/pdf-craft/meter"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

const (
	hintingNone = 0
	notdef      = 0
)

type Font struct {
	alias          string
	face           *sfnt.Font
	rawData        []byte
	compressedData []byte
	manager        fontManager
	metrics        Metrics
}

type Glyph struct {
	rune    uint16
	index   sfnt.GlyphIndex
	advance fixed.Int26_6
}

// Metrics содержит параметры шрифта. Визуальное представление можно найти здесь:
// https://developer.apple.com/library/mac/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyph_metrics_2x.png
//
// Ниже представлены значения для LiberationSans-Regular.ttf
// /Flags 4
// /FontBBox [-203 -303 1050 910]
// /ItalicAngle 0
// /Ascent 905
// /Descent -212
// /CapHeight 688
// /StemV 569
type Metrics struct {
	Ascent    int
	Descent   int
	CapHeight int
	StemV     int
	Ppem      fixed.Int26_6
	FontBBox  []int
}

type fontManager struct {
	glyphsCache map[rune]*Glyph
	glyphs      []*Glyph
	dirtyFlag   bool
}

func defaultGlyph() *Glyph {
	return &Glyph{
		rune:    '□',
		index:   0,
		advance: 600,
	}
}

func NewFont(path, alias string) (*Font, error) {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	sfntFont, err := sfnt.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	ppem := fixed.I(1000)

	var buf sfnt.Buffer

	sfntMetrics, err := sfntFont.Metrics(&buf, ppem, hintingNone)
	if err != nil {
		return nil, err
	}

	ascent := sfntMetrics.Ascent.Round()
	descent := sfntMetrics.Descent.Round()
	capHeight := sfntMetrics.CapHeight.Round()

	bounds, err := sfntFont.Bounds(&buf, ppem, hintingNone)
	if err != nil {
		return nil, err
	}

	// [Min.X, Min.Y, Max.X, Max.Y]
	// Min.Y и Max.Y поменяны местами нарочно ввиду того, что у PDF начало координат находится в левом нижнем углу.
	fontBBox := []int{
		bounds.Min.X.Round(),
		-bounds.Max.Y.Round(),
		bounds.Max.X.Round(),
		-bounds.Min.Y.Round(),
	}

	// Среднее значение
	glyphAdvanceRounded := 80

	glyphIndex, err := sfntFont.GlyphIndex(&buf, 'I')
	if err != nil {
		return nil, err
	}

	glyphAdvance, err := sfntFont.GlyphAdvance(&buf, glyphIndex, ppem, hintingNone)
	if err != nil {
		return nil, err
	}

	glyphAdvanceRounded = glyphAdvance.Round()
	if glyphAdvanceRounded > 20 && glyphAdvanceRounded < 200 {
		glyphAdvanceRounded = glyphAdvanceRounded * 3 / 4
	}

	stemV := glyphAdvanceRounded

	metrics := Metrics{
		Ascent:    ascent,
		Descent:   descent,
		CapHeight: capHeight,
		StemV:     stemV,
		Ppem:      ppem,
		FontBBox:  fontBBox,
	}

	f := &Font{
		alias:   alias,
		face:    sfntFont,
		rawData: fontBytes,
		metrics: metrics,
		manager: fontManager{
			glyphsCache: make(map[rune]*Glyph),
		},
	}

	return f, nil
}

func (f *Font) Alias() string {
	return f.alias
}

func (f *Font) GID(r rune) uint16 {
	gl, ok := f.manager.glyphsCache[r]
	if !ok {
		return notdef
	}

	return uint16(gl.index)
}

func (f *Font) Glyphs() []*Glyph {
	if !f.manager.dirtyFlag {
		return f.manager.glyphs
	}

	glyphs := slices.SortedFunc(maps.Values(f.manager.glyphsCache), func(g *Glyph, g2 *Glyph) int {
		return cmp.Compare(g.index, g2.index)
	})

	f.manager.glyphs = glyphs

	return glyphs
}

func (f *Font) Metrics() Metrics {
	return f.metrics
}

func (f *Font) Data() (data []byte, uncompressedLen int, err error) {
	uncompressedLen = len(f.rawData)

	if !f.manager.dirtyFlag {
		return f.compressedData, uncompressedLen, nil
	}

	subset, err := f.ttfSubset()
	if err != nil {
		return nil, 0, err
	}

	compressedData, err := compress(subset)
	if err != nil {
		return nil, 0, err
	}

	f.compressedData = compressedData
	//f.compressedData = subset
	f.manager.dirtyFlag = false

	return compressedData, uncompressedLen, nil
}

func compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)

	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}

	w.Close()

	return b.Bytes(), nil
}

func (f *Font) SaveRunes(text string) {
	for _, r := range text {
		_, ok := f.manager.glyphsCache[r]
		if ok {
			continue
		}

		if r == '\n' {
			continue
		}

		buf := new(sfnt.Buffer)

		index, _ := f.face.GlyphIndex(buf, r) //err is always nil

		adv, err := f.face.GlyphAdvance(buf, index, f.metrics.Ppem, hintingNone)
		if err != nil {
			adv = 600
		}

		f.manager.addGlyph(r, index, adv)

		f.manager.dirtyFlag = true
	}
}

func (f *Font) MeasureText(fontSize meter.PT, text string) meter.PT {
	var (
		advance fixed.Int26_6
		buf     sfnt.Buffer
	)

	fontSizeEm := fontSize.FixedI()
	ppemFont := f.metrics.Ppem.Mul(fontSizeEm)
	prevGlyphIndex := sfnt.GlyphIndex(0)

	for _, c := range text {
		gl, ok := f.manager.glyphsCache[c]
		if !ok {
			gl = defaultGlyph()
		}

		advance += gl.advance.Mul(fontSizeEm)

		if prevGlyphIndex > 0 {
			kern, err := f.face.Kern(&buf, prevGlyphIndex, gl.index, ppemFont, hintingNone)
			if err != nil {
				//TODO: обработка ошибок
			}

			advance += kern
		}

		prevGlyphIndex = gl.index
	}

	width := meter.PT((advance / fixed.Int26_6(f.metrics.Ppem.Round())).Round())

	return width
}

func (f *Font) SplitText(text string, size meter.PT, width meter.MM) []string {
	lines := make([]string, 0)

	for seg := range strings.Lines(text) {
		lines = append(lines, f.splitSegment(seg, size, width)...)
	}

	return lines
}

func (f *Font) splitSegment(text string, size meter.PT, width meter.MM) []string {
	words := strings.Fields(text)

	lines := make([]string, 0, len(words))

	line := words[0]

	for _, word := range words[1:] {
		candidate := line + " " + word

		//TODO: хранить text_width, чтобы не считать заново при позиционировании

		candidateWidth := f.MeasureText(size, candidate)
		if candidateWidth > width.PT() {
			lines = append(lines, line)

			line = word

			continue
		}

		line = candidate
	}

	lines = append(lines, line)

	return lines
}

func (g *Glyph) Index() uint16 {
	return uint16(g.index)
}

func (g *Glyph) Rune() uint16 {
	return g.rune
}

func (g *Glyph) Advance() int64 {
	return int64(g.advance.Round())
}

func (mgr *fontManager) addGlyph(r rune, index sfnt.GlyphIndex, advance fixed.Int26_6) {
	mgr.glyphsCache[r] = &Glyph{
		rune:    uint16(r),
		index:   index,
		advance: advance,
	}
}
