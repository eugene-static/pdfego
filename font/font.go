package font

import (
	"bytes"
	"cmp"
	"compress/zlib"
	"maps"
	"os"
	"slices"

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
	textBuffer     []rune
	glyphsCache    map[rune]Glyph
	glyphFastCache []Glyph
	glyphs         []Glyph
	dirtyFlag      bool
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

	//fontFile, err := os.Open(path)
	//if err != nil {
	//	return nil, err
	//}
	//
	//defer fontFile.Close()
	//
	//face, err := font.ParseTTF(fontFile)
	//if err != nil {
	//	return nil, err
	//}
	//
	//input := shaping.Input{
	//	RunStart:     0,
	//	Direction:    di.DirectionLTR,
	//	Face:         face,
	//	FontFeatures: nil,
	//}

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
			glyphsCache:    make(map[rune]Glyph),
			glyphFastCache: make([]Glyph, 1200),
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

func (f *Font) Glyphs() []Glyph {
	if !f.manager.dirtyFlag {
		return f.manager.glyphs
	}

	glyphs := slices.SortedFunc(maps.Values(f.manager.glyphsCache), func(g Glyph, g2 Glyph) int {
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

func (f *Font) saveRune(r rune) {
	if r < 1200 && f.manager.glyphFastCache[r].rune > 0 {
		return
	}

	_, ok := f.manager.glyphsCache[r]
	if ok {
		return
	}

	if r == '\n' {
		return
	}

	buf := new(sfnt.Buffer)

	gid, _ := f.face.GlyphIndex(buf, r) //err is always nil

	adv, err := f.face.GlyphAdvance(buf, gid, f.metrics.Ppem, hintingNone)
	if err != nil {
		adv = 600
	}

	f.manager.addGlyph(r, gid, adv)

	f.manager.dirtyFlag = true
}

func (f *Font) MeasureText(fontSize meter.PT, text []rune, start, end int) meter.PT {
	if start < 0 || end > len(text) || start >= end {
		return meter.PT(0)
	}

	var advance fixed.Int26_6

	for i := start; i < end; i++ {
		char := text[i]

		if char < 1200 {
			gl := f.manager.glyphFastCache[char]

			if gl.advance > 0 {
				advance += gl.advance
				continue
			}
		}

		gl, ok := f.manager.glyphsCache[text[i]]
		if !ok {
			advance += 600
			continue
		}

		advance += gl.advance
	}

	fontSizeEm := fontSize.FixedI()
	advance = advance.Mul(fontSizeEm)

	width := meter.PT(float64(advance) / float64(f.metrics.Ppem))

	return width
}

func (f *Font) FullText(text string, size meter.PT) Segment {
	f.manager.textBuffer = f.manager.textBuffer[:0]

	for _, r := range text {
		f.manager.textBuffer = append(f.manager.textBuffer, r)
		f.saveRune(r)
	}

	textWidth := f.MeasureText(size, f.manager.textBuffer, 0, len(text))

	return Segment{
		text:  text,
		width: textWidth.MM(),
	}
}

func (f *Font) SplitText(text string, size meter.PT, width meter.MM, buf []Segment) []Segment {
	segments := buf[:0]
	f.manager.textBuffer = f.manager.textBuffer[:0]
	targetWidth := width.PT()

	for _, r := range text {
		f.manager.textBuffer = append(f.manager.textBuffer, r)
		f.saveRune(r)
	}

	start := 0
	for start < len(f.manager.textBuffer) && !f.manager.wrapSymbols(start) {
		start++
	}

	lineStart := 0
	lineEnd := start
	textWidth := f.MeasureText(size, f.manager.textBuffer, lineStart, lineEnd)

	if start >= len(f.manager.textBuffer) {
		segments = append(segments, Segment{
			text:  string(f.manager.textBuffer),
			width: textWidth.MM(),
		})

		return segments
	}

	if cap(segments) < 10 {
		segments = slices.Grow(segments, 10)
	}

	for i := start; i < len(f.manager.textBuffer); {
		for i < len(f.manager.textBuffer) && f.manager.wrapSymbols(i) {
			i++
		}

		wordStart := i

		for i < len(f.manager.textBuffer) && !f.manager.wrapSymbols(i) {
			i++
		}

		wordEnd := i

		candidateWidth := f.MeasureText(size, f.manager.textBuffer, lineStart, wordEnd)

		if candidateWidth > targetWidth || f.manager.textBuffer[lineEnd] == '\n' {
			segments = append(segments, Segment{
				text:  string(f.manager.textBuffer[lineStart:lineEnd]),
				width: textWidth.MM(),
			})

			lineStart = wordStart
			lineEnd = wordEnd
			textWidth = candidateWidth - textWidth

			continue
		}

		lineEnd = wordEnd
		textWidth = candidateWidth
	}

	if lineStart < len(f.manager.textBuffer) {
		lineWidth := f.MeasureText(size, f.manager.textBuffer, lineStart, lineEnd)

		segments = append(segments, Segment{
			text:  string(f.manager.textBuffer[lineStart:lineEnd]),
			width: lineWidth.MM(),
		})
	}

	return segments
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

func (mgr *fontManager) addGlyph(r rune, gid sfnt.GlyphIndex, advance fixed.Int26_6) {
	if r < 1200 {
		mgr.glyphFastCache[r].rune = uint16(r)
		mgr.glyphFastCache[r].index = gid
		mgr.glyphFastCache[r].advance = advance
		mgr.glyphsCache[r] = mgr.glyphFastCache[r]

		return
	}

	mgr.glyphsCache[r] = Glyph{
		rune:    uint16(r),
		index:   gid,
		advance: advance,
	}
}

func (mgr *fontManager) wrapSymbols(index int) bool {
	if index >= len(mgr.textBuffer) {
		return false
	}

	return mgr.textBuffer[index] == ' ' || mgr.textBuffer[index] == '\t' || mgr.textBuffer[index] == '\n' || mgr.textBuffer[index] == '\r'
}
