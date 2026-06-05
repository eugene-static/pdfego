package font

import (
	"cmp"
	"maps"
	"os"
	"slices"

	"github.com/eugene-static/pdf-craft/pkg/unit"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

const (
	hintingNone               = 0
	notdef                    = 0
	ppem        fixed.Int26_6 = 1000 << 6

	splitTab     = '\t'
	splitNewline = '\n'
	splitReturn  = '\r'
	splitSpace   = ' '

	defaultAdvance     fixed.Int26_6 = 600
	rusRunesLimitIndex               = 1200
)

type Font struct {
	alias          string
	face           *sfnt.Font
	rawData        []byte
	compressedData []byte
	manager        fontManager
	metrics        Metrics
}

// Metrics содержит параметры шрифта. Визуальное представление можно найти здесь:
// https://developer.apple.com/library/mac/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyph_metrics_2x.png
//
// Ниже представлены значения для LiberationSans-Regular.ttf:
//
//	/Flags 4
//	/FontBBox [-203 -303 1050 910]
//	/ItalicAngle 0
//	/Ascent 905
//	/Descent -212
//	/CapHeight 688
//	/StemV 80
type Metrics struct {
	FontBBox  []int
	Ascent    int
	Descent   int
	CapHeight fixed.Int26_6
	StemV     int
	Height    fixed.Int26_6
}

func New(path, alias string) (*Font, error) {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	sfntFont, err := sfnt.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	buf := new(sfnt.Buffer)

	sfntMetrics, err := sfntFont.Metrics(buf, ppem, hintingNone)
	if err != nil {
		return nil, err
	}

	height := sfntMetrics.Height
	capHeight := sfntMetrics.CapHeight
	ascent := sfntMetrics.Ascent.Round()
	descent := sfntMetrics.Descent.Round()

	bounds, err := sfntFont.Bounds(buf, ppem, hintingNone)
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

	glyphIndex, err := sfntFont.GlyphIndex(buf, 'I')
	if err != nil {
		return nil, err
	}

	glyphAdvance, err := sfntFont.GlyphAdvance(buf, glyphIndex, ppem, hintingNone)
	if err != nil {
		return nil, err
	}

	glyphAdvanceRounded = glyphAdvance.Round()
	if glyphAdvanceRounded > 20 && glyphAdvanceRounded < 200 {
		glyphAdvanceRounded = glyphAdvanceRounded * 3 / 4
	}

	stemV := glyphAdvanceRounded

	metrics := Metrics{
		FontBBox:  fontBBox,
		Ascent:    ascent,
		Descent:   descent,
		CapHeight: capHeight,
		StemV:     stemV,
		Height:    height,
	}

	f := &Font{
		alias:   alias,
		face:    sfntFont,
		rawData: fontBytes,
		metrics: metrics,
		manager: fontManager{
			faceBuffer:     buf,
			glyphsCache:    make(map[rune]Glyph),
			glyphFastCache: make([]Glyph, rusRunesLimitIndex),
		},
	}

	return f, nil
}

func (f *Font) Alias() string {
	return f.alias
}

func (f *Font) Height(size unit.PT) unit.PT {
	return size * unit.PT(float64(f.metrics.Height)/float64(ppem))
}

func (f *Font) CapHeight(size unit.PT) unit.PT {
	return size * unit.PT(float64(f.metrics.CapHeight)/float64(ppem))
}

func (f *Font) GID(r rune) uint16 {
	if r < rusRunesLimitIndex {
		gl := f.manager.glyphFastCache[r]

		return gl.Index()
	}

	gl, ok := f.manager.glyphsCache[r]
	if ok {
		return gl.Index()
	}

	return notdef
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

func (f *Font) Bytes() []byte {
	return f.rawData
}

func (f *Font) CompressedBytes() ([]byte, bool) {
	return f.compressedData, !f.manager.dirtyFlag
}

func (f *Font) Subset() ([]byte, error) {
	subset, err := f.ttfSubset()
	if err != nil {
		return nil, err
	}

	return subset, nil
}

func (f *Font) SaveCompressedBytes(data []byte) {
	f.compressedData = data

	f.manager.dirtyFlag = false
}

func (f *Font) MeasureText(fontSize unit.PT, text []rune, start, end int) unit.PT {
	if start < 0 || end > len(text) || start >= end {
		return unit.PT(0)
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
			advance += defaultAdvance
			continue
		}

		advance += gl.advance
	}

	fontSizeEm := fontSize.FixedI()
	advance = advance.Mul(fontSizeEm)

	width := unit.PT(float64(advance) / float64(ppem))

	return width
}

func (f *Font) FullText(text string, size unit.PT) Text {
	f.manager.textBuffer = f.manager.textBuffer[:0]

	for _, r := range text {
		f.manager.textBuffer = append(f.manager.textBuffer, r)
		f.saveRune(r)
	}

	textWidth := f.MeasureText(size, f.manager.textBuffer, 0, len(text))

	return Text{
		data:  text,
		width: textWidth.MM(),
	}
}

func (f *Font) SplitText(text string, size unit.PT, width unit.MM, buf []Text) []Text {
	segments := buf[:0]
	f.manager.textBuffer = f.manager.textBuffer[:0]

	targetWidth := width.PT()

	for _, r := range text {
		f.manager.textBuffer = append(f.manager.textBuffer, r)
		f.saveRune(r)
	}

	start := 0
	for start < len(f.manager.textBuffer) && (!f.manager.wrapSymbols(start) || targetWidth == 0) {
		start++
	}

	lineStart := 0
	lineEnd := start
	textWidth := f.MeasureText(size, f.manager.textBuffer, lineStart, lineEnd)

	if start >= len(f.manager.textBuffer) {
		segments = append(segments, Text{
			data:  string(f.manager.textBuffer),
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

		if candidateWidth > targetWidth || f.manager.textBuffer[lineEnd] == splitNewline {
			segments = append(segments, Text{
				data:  string(f.manager.textBuffer[lineStart:lineEnd]),
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

		segments = append(segments, Text{
			data:  string(f.manager.textBuffer[lineStart:lineEnd]),
			width: lineWidth.MM(),
		})
	}

	return segments
}

func (f *Font) saveRune(r rune) {
	if r < rusRunesLimitIndex && f.manager.glyphFastCache[r].rune > 0 {
		return
	}

	_, ok := f.manager.glyphsCache[r]
	if ok {
		return
	}

	if r == splitNewline {
		return
	}

	gid, _ := f.face.GlyphIndex(f.manager.faceBuffer, r) // err всегда nil

	advance, err := f.face.GlyphAdvance(
		f.manager.faceBuffer,
		gid,
		ppem,
		hintingNone,
	)
	if err != nil {
		advance = defaultAdvance // нас не интересует ошибка, просто ставим среднюю ширину символа.
	}

	f.manager.addGlyph(r, gid, advance)

	f.manager.dirtyFlag = true
}

type Glyph struct {
	rune    uint16
	index   sfnt.GlyphIndex
	advance fixed.Int26_6
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

type fontManager struct {
	faceBuffer     *sfnt.Buffer
	textBuffer     []rune
	glyphs         []Glyph
	glyphFastCache []Glyph
	glyphsCache    map[rune]Glyph
	dirtyFlag      bool
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

	return mgr.textBuffer[index] == splitSpace ||
		mgr.textBuffer[index] == splitTab ||
		mgr.textBuffer[index] == splitNewline ||
		mgr.textBuffer[index] == splitReturn
}

type Text struct {
	data  string
	width unit.MM
}

func (s *Text) Data() string {
	return s.data
}

func (s *Text) Width() unit.MM {
	return s.width
}

func Padding(fontSize unit.PT) unit.PT {
	return fontSize / 7
}
