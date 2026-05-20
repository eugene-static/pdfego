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
	"github.com/go-text/typesetting/shaping"
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
	shaper         shaper
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

type shaper struct {
	shaper shaping.Shaper
	input  shaping.Input
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

		gid, _ := f.face.GlyphIndex(buf, r) //err is always nil

		adv, err := f.face.GlyphAdvance(buf, gid, f.metrics.Ppem, hintingNone)
		if err != nil {
			adv = 600
		}

		f.manager.addGlyph(r, gid, adv)

		f.manager.dirtyFlag = true
	}
}

func (f *Font) MeasureText(fontSize meter.PT, text string) meter.PT {
	var (
		advance fixed.Int26_6
	)

	fontSizeEm := fontSize.FixedI()

	for _, c := range text {
		gl, ok := f.manager.glyphsCache[c]
		if !ok {
			gl = defaultGlyph()
		}

		advance += gl.advance.Mul(fontSizeEm)
	}

	width := meter.PT(float64(advance) / float64(f.metrics.Ppem))

	return width
}

func (f *Font) SplitText(text string, size meter.PT, width meter.MM) []Segment {
	lines := make([]Segment, 0)

	for seg := range strings.Lines(text) {
		lines = slices.Concat(lines, f.splitSegment(seg, size, width))
	}

	return lines
}

func (f *Font) SplitTextOptimized(buf []Segment, text string, size meter.PT, width meter.MM) (segments []Segment) {
	copy(segments, buf[:0])
	targetWidth := width.PT()

	start := 0
	for start < len(text) && (text[start] == ' ' || text[start] == '\t' || text[start] == '\n' || text[start] == '\r') {
		start++
	}

	if start >= len(text) {
		return segments
	}

	lineStart := 0
	lineEnd := start
	textWidth := f.MeasureText(size, text[:lineEnd])

	for i := start; i < len(text); {
		for i < len(text) && (text[i] == ' ' || text[i] == '\t' || text[i] == '\n' || text[i] == '\r') {
			i++
		}

		wordStart := i
		for i < len(text) && !(text[i] == ' ' || text[i] == '\t' || text[i] == '\n' || text[i] == '\r') {
			i++
		}

		wordEnd := i

		candidate := text[lineStart:wordEnd]
		candidateWidth := f.MeasureText(size, candidate)

		if candidateWidth > targetWidth || text[lineEnd] == '\n' {
			currentLine := strings.Clone(text[lineStart:lineEnd])

			segments = append(segments, Segment{
				text:  currentLine,
				width: textWidth.MM(),
			})

			lineStart = wordStart
			lineEnd = wordEnd

			continue
		}

		lineEnd = wordEnd
		textWidth = candidateWidth
	}

	if lineStart < len(text) {
		lastLine := strings.Clone(text[lineStart:lineEnd])
		lineWidth := f.MeasureText(size, lastLine)

		segments = append(segments, Segment{
			text:  lastLine,
			width: lineWidth.MM(),
		})
	}

	return segments
}

func (f *Font) splitSegment(text string, size meter.PT, width meter.MM) []Segment {
	words := strings.Fields(text)

	segments := make([]Segment, 0, len(words))

	line := words[0]
	textWidth := meter.PT(0)

	for _, word := range words[1:] {
		candidate := line + " " + word

		candidateWidth := f.MeasureText(size, candidate)
		if candidateWidth > width.PT() {
			segments = append(segments, Segment{
				text:  line,
				width: textWidth.MM(),
			})

			line = word

			continue
		}

		textWidth = candidateWidth
		line = candidate
	}

	segments = append(segments, Segment{
		text:  line,
		width: textWidth.MM(),
	})

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
	mgr.glyphsCache[r] = &Glyph{
		rune:    uint16(r),
		index:   gid,
		advance: advance,
	}
}
