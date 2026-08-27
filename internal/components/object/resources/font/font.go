package font

import (
	"cmp"
	"maps"
	"os"
	"slices"
	"sync"

	"github.com/eugene-static/pdfego/internal/components/parameter"
	"github.com/eugene-static/pdfego/internal/components/stream"
	"github.com/eugene-static/pdfego/internal/components/stream/primitives"
	"github.com/eugene-static/pdfego/internal/compressor"
	"github.com/eugene-static/pdfego/unit"
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
	alias   primitives.Alias
	face    *sfnt.Font
	bytes   []byte
	manager fontManager
	metrics metrics
	objects objects
}

// metrics содержит параметры шрифта. Визуальное представление можно найти здесь:
// https://developer.apple.com/library/mac/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyph_metrics_2x.png
type metrics struct {
	fontBBox          []int
	italicAngle       float64
	ascent            int
	descent           int
	stemV             int
	underlinePosition fixed.Int26_6
	capHeight         fixed.Int26_6
	height            fixed.Int26_6
}

func New(path, alias string, comp *compressor.Compressor) (*Font, error) {
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

	italicAngle := float64(0)
	underlinePosition := int16(0)

	if postTable := sfntFont.PostTable(); postTable != nil {
		italicAngle = postTable.ItalicAngle
		underlinePosition = postTable.UnderlinePosition
	}

	_metrics := metrics{
		fontBBox:          fontBBox,
		italicAngle:       italicAngle,
		ascent:            ascent,
		descent:           descent,
		capHeight:         capHeight,
		underlinePosition: fixed.I(int(underlinePosition)),
		stemV:             stemV,
		height:            height,
	}

	f := &Font{
		alias:   primitives.Alias(alias),
		face:    sfntFont,
		bytes:   fontBytes,
		metrics: _metrics,
		manager: fontManager{
			mu:              &sync.RWMutex{},
			faceb:           buf,
			glyphsSlowCache: make(map[rune]glyph),
			glyphsFastCache: make([]glyph, rusRunesLimitIndex),
			compressor:      comp,
		},
	}

	return f, nil
}

// 1 0 0 1 $X $Y Tm <$HEX1$HEX2...$HEXN> Tj
func (f *Font) WriteToStream(dst *stream.Stream, text []Text, fontSize unit.PT) {
	builder := dst.NewStreamWriter()

	builder.
		Write(primitives.BeginText).LF().
		Write(f.alias).SP().
		Write(fontSize).SP().
		Write(primitives.TextFont).LF()

	for i := range text {
		builder.
			Write(text[i].matrix).SP().
			Write(primitives.TextMatrix).SP().
			Write(text[i].content).SP().
			Write(primitives.ShowText).LF()
	}

	builder.
		Write(primitives.EndText).LF().
		Close()
}

func (f *Font) Alias() parameter.Name {
	return parameter.Name(f.alias)
}

func (f *Font) Height(size unit.PT) unit.PT {
	return size * unit.PT(float64(f.metrics.height)/float64(ppem))
}

func (f *Font) CapHeight(size unit.PT) unit.PT {
	return size * unit.PT(float64(f.metrics.capHeight)/float64(ppem))
}

func (f *Font) UnderlinePosition(size unit.PT) unit.PT {
	return size * unit.PT(float64(f.metrics.underlinePosition)/float64(ppem)) / 7
}

func (f *Font) SplitText(text string, size unit.PT, width unit.MM, textb *[]Text, runeb *[]rune, hexb *[]primitives.HEX) {
	*runeb = (*runeb)[:0]

	targetWidth := width.PT()

	for _, r := range text {
		f.saveRune(r)
		*runeb = append(*runeb, r)
	}

	start := 0
	for start < len(*runeb) && (!isSpaceSymbol((*runeb)[start]) || targetWidth == 0) {
		start++
	}

	lineStart := 0
	lineEnd := start
	textWidth := f.manager.measureText(*runeb, size, lineStart, lineEnd)

	if start >= len(*runeb) {
		*textb = append(*textb, Text{
			content: f.text(*runeb, hexb),
			width:   textWidth.MM(),
		})

		return
	}

	for i := start; i < len(*runeb); {
		for i < len(*runeb) && isSpaceSymbol((*runeb)[i]) {
			i++
		}

		wordStart := i

		for i < len(*runeb) && !isSpaceSymbol((*runeb)[i]) {
			i++
		}

		wordEnd := i

		candidateWidth := f.manager.measureText(*runeb, size, lineStart, wordEnd)

		if candidateWidth > targetWidth || (*runeb)[lineEnd] == splitNewline {
			*textb = append(*textb, Text{
				content: f.text((*runeb)[lineStart:lineEnd], hexb),
				width:   textWidth.MM(),
			})

			textWidth = candidateWidth - textWidth
			if (*runeb)[lineEnd] == splitSpace {
				textWidth -= f.manager.glyphWidth(splitSpace, size)
			}

			lineStart = wordStart
			lineEnd = wordEnd

			continue
		}

		lineEnd = wordEnd
		textWidth = candidateWidth
	}

	if lineStart < len(*runeb) {
		lineWidth := f.manager.measureText(*runeb, size, lineStart, lineEnd)

		*textb = append(*textb, Text{
			content: f.text((*runeb)[lineStart:lineEnd], hexb),
			width:   lineWidth.MM(),
		})
	}
}

func (f *Font) gid(r rune) primitives.HEX {
	gl, ok := f.manager.glyph(r)
	if ok {
		return primitives.HEX(gl.index)
	}

	return notdef
}

func (f *Font) saveRune(r rune) {
	gl, ok := f.manager.glyph(r)
	if ok && gl.rune > 0 {
		return
	}

	if r == splitNewline {
		return
	}

	gid, _ := f.face.GlyphIndex(f.manager.faceb, r) // err всегда nil

	advance, err := f.face.GlyphAdvance(f.manager.faceb, gid, ppem, hintingNone)
	if err != nil {
		advance = defaultAdvance // нас не интересует ошибка, просто ставим среднюю ширину символа.
	}

	f.manager.addGlyph(r, gid, advance)

	f.manager.dirtyFlag = true
}

func (f *Font) glyphs() []glyph {
	if !f.manager.dirtyFlag {
		return f.manager.glyphs
	}

	glyphs := slices.SortedFunc(maps.Values(f.manager.glyphsSlowCache), func(g glyph, g2 glyph) int {
		return cmp.Compare(g.index, g2.index)
	})

	f.manager.glyphs = glyphs

	return glyphs
}

func (f *Font) glyphsWidthTable() parameter.GlyphsWidthTable {
	glyphs := f.glyphs()

	widthTable := make(parameter.GlyphsWidthTable, 0, len(glyphs))

	for _, gl := range glyphs {
		widthTable = append(widthTable,
			[2]uint64{
				uint64(gl.index),
				uint64(gl.advance.Round()),
			})
	}

	return widthTable
}

func (f *Font) subset(str *stream.Stream) error {
	subset, err := f.ttfSubset()
	if err != nil {
		return err
	}

	str.Reset()

	bytes, err := f.manager.compressor.ForceCompress(subset)
	if err != nil {
		return err
	}

	str.Write(bytes)

	return nil
}

func (f *Font) text(s []rune, hexb *[]primitives.HEX) primitives.String {
	start := len(*hexb)
	end := start

	for _, r := range s {
		*hexb = append(*hexb, f.gid(r))
		end++
	}

	return (*hexb)[start:end]
}

type glyph struct {
	rune    uint16
	index   sfnt.GlyphIndex
	advance fixed.Int26_6
}

type fontManager struct {
	mu              *sync.RWMutex
	compressor      *compressor.Compressor
	faceb           *sfnt.Buffer
	glyphs          []glyph
	glyphsFastCache []glyph
	glyphsSlowCache map[rune]glyph
	dirtyFlag       bool
}

func (mgr *fontManager) measureText(runeb []rune, fontSize unit.PT, start, end int) unit.PT {
	if start < 0 || end > len(runeb) || start >= end {
		return unit.PT(0)
	}

	var advance fixed.Int26_6

	for i := start; i < end; i++ {
		char := runeb[i]

		gl, ok := mgr.glyph(char)
		if !ok {
			advance += defaultAdvance

			continue
		}

		advance += gl.advance
	}

	advance = advance.Mul(fontSize.FixedI())

	width := unit.PT(float64(advance) / float64(ppem))

	return width
}

func (mgr *fontManager) addGlyph(r rune, gid sfnt.GlyphIndex, advance fixed.Int26_6) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	if r < rusRunesLimitIndex {
		mgr.glyphsFastCache[r].rune = uint16(r)
		mgr.glyphsFastCache[r].index = gid
		mgr.glyphsFastCache[r].advance = advance
		mgr.glyphsSlowCache[r] = mgr.glyphsFastCache[r]

		return
	}

	mgr.glyphsSlowCache[r] = glyph{
		rune:    uint16(r),
		index:   gid,
		advance: advance,
	}
}

func (mgr *fontManager) glyph(r rune) (glyph, bool) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	if r < rusRunesLimitIndex {
		return mgr.glyphsFastCache[r], true
	}

	gl, ok := mgr.glyphsSlowCache[r]

	return gl, ok
}

func (mgr *fontManager) glyphWidth(r rune, size unit.PT) unit.PT {
	gl, ok := mgr.glyph(r)
	if !ok {
		return 0
	}

	fontSizeEm := size.FixedI()
	advance := gl.advance.Mul(fontSizeEm)

	width := unit.PT(float64(advance) / float64(ppem))

	return width
}

func isSpaceSymbol(r rune) bool {
	return r == splitSpace || r == splitTab || r == splitNewline || r == splitReturn
}

type Text struct {
	content primitives.String
	matrix  primitives.Matrix
	width   unit.MM
	dw      unit.MM // TODO: Разница между шириной ячейки и шириной текста
}

func (text *Text) SetMatrix(point primitives.Point) {
	matrix := primitives.NewDefaultTextMatrix(point)

	text.matrix = matrix
}

func (text *Text) Width() unit.MM {
	return text.width
}

// Padding -- расстояние от текста до границ объекта, в котором он расположен.
func Padding(fontSize unit.PT) unit.PT {
	return fontSize / 7
}

func UnderlineSize(fontSize unit.PT) unit.PT {
	return fontSize / 30
}
