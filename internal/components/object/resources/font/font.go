package font

import (
	"cmp"
	"maps"
	"os"
	"slices"
	"sync"
	"unicode"

	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"

	"github.com/eugene-static/pdfego/internal/components/parameter"
	"github.com/eugene-static/pdfego/internal/components/stream"
	"github.com/eugene-static/pdfego/internal/components/stream/bytes"
	"github.com/eugene-static/pdfego/internal/components/stream/primitives"
	"github.com/eugene-static/pdfego/internal/compressor"
	"github.com/eugene-static/pdfego/unit"
)

const (
	hintingNone        = 0
	rusRunesLimitIndex = 1200

	ppem           fixed.Int26_6 = 1000 << 6
	defaultAdvance fixed.Int26_6 = 600

	lf = '\n'
)

type Font struct {
	alias   primitives.Alias
	face    *sfnt.Font
	bytes   []byte
	manager fontManager
	metrics metrics
	objects objects
}

// Структура metrics содержит параметры шрифта. Визуальное представление можно найти здесь:
// https://developer.apple.com/library/mac/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyph_metrics_2x.png
type metrics struct {
	fontBBox          []int
	italicAngle       float64
	ascent            int
	descent           int
	stemV             int
	height            unit.EM
	capHeight         unit.EM
	underlinePosition unit.EM
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

	var (
		italicAngle       float64
		underlinePosition int16
	)

	if postTable := sfntFont.PostTable(); postTable != nil {
		italicAngle = postTable.ItalicAngle
		underlinePosition = postTable.UnderlinePosition
	}

	_metrics := metrics{
		fontBBox:          fontBBox,
		italicAngle:       italicAngle,
		ascent:            ascent,
		descent:           descent,
		stemV:             stemV,
		height:            unit.Font(height).EM(ppem),
		capHeight:         unit.Font(capHeight).EM(ppem),
		underlinePosition: unit.Font(fixed.I(int(underlinePosition))).EM(ppem),
	}

	_fontManager := fontManager{
		mu:                 &sync.RWMutex{},
		faceb:              buf,
		glyphsSlowCache:    make(map[rune]glyph),
		glyphsFastCache:    make([]glyph, rusRunesLimitIndex),
		spaceGlyphsIndexes: make(map[unit.HEX]struct{}, 4),
		compressor:         comp,
	}

	f := &Font{
		alias:   primitives.Alias(alias),
		face:    sfntFont,
		bytes:   fontBytes,
		metrics: _metrics,
		manager: _fontManager,
	}

	return f, nil
}

func (f *Font) WriteToStreamWithIndividualGlyphPosition(dst *stream.Stream, text []Text, fontSize unit.PT) {
	bw := bytes.NewWriter(dst.AvailableBuffer())

	bw.
		Write(primitives.BeginText).LF().
		Write(f.alias).SP().
		Write(fontSize).SP().
		Write(primitives.TextFont).LF()

	for i := range text {
		bw.
			Write(text[i].matrix).SP().
			Write(primitives.TextMatrix).SP().
			WriteByte(primitives.ArrayOpen).
			WriteByte(primitives.HexadecimalStringOpen)

		for _, hex := range text[i].content {
			bw.Write(hex)

			if f.manager.isSpaceGlyphIndex(hex) {
				bw.
					WriteByte(primitives.HexadecimalStringClose).SP().
					WriteFloat(float64(text[i].shift.Font(ppem))).SP().
					WriteByte(primitives.HexadecimalStringOpen)
			}
		}

		bw.
			WriteByte(primitives.HexadecimalStringClose).
			WriteByte(primitives.ArrayClose).SP().
			Write(primitives.ShowTextWithIndividualGlyphPositioning).LF()
	}

	bw.Write(primitives.EndText).LF()

	dst.Write(bw.Bytes())
}

func (f *Font) WriteToStream(dst *stream.Stream, text []Text, fontSize unit.PT) {
	sw := dst.NewStreamWriter()

	sw.
		Write(primitives.BeginText).LF().
		Write(f.alias).SP().
		Write(fontSize).SP().
		Write(primitives.TextFont).LF()

	for i := range text {
		sw.
			Write(text[i].matrix).SP().
			Write(primitives.TextMatrix).SP().
			Write(text[i].content).SP().
			Write(primitives.ShowText).LF()
	}

	sw.
		Write(primitives.EndText).LF().
		Close()
}

func (f *Font) Alias() parameter.Name {
	return parameter.Name(f.alias)
}

func (f *Font) Height(size unit.PT) unit.PT {
	return f.metrics.height.PT(size)
}

func (f *Font) CapHeight(size unit.PT) unit.PT {
	return f.metrics.capHeight.PT(size)
}

func (f *Font) UnderlinePosition(size unit.PT) unit.PT {
	return f.metrics.underlinePosition.PT(size) / 7
}

func (f *Font) WriteTextToLine(text string, size unit.PT, width unit.MM, textb []Text, hexb *[]unit.HEX) []Text {
	var (
		shift       unit.EM
		lineAdvance unit.EM
		spaceCount  int
	)

	targetAdvance := width.PT().EM(size)
	hexbStartIndex := len(*hexb)

	for _, r := range text {
		_glyph := f.glyph(r)

		if unicode.IsSpace(r) {
			spaceCount++

			f.manager.saveSpaceGlyphIndex(_glyph.index)
		}

		lineAdvance += _glyph.advance
		*hexb = append(*hexb, _glyph.index)
	}

	if spaceCount > 0 {
		shift = (lineAdvance - targetAdvance) / unit.EM(spaceCount)
	}

	textb = append(textb, Text{
		content: (*hexb)[hexbStartIndex:],
		width:   lineAdvance.PT(size).MM(),
		shift:   shift,
	})

	return textb
}

func (f *Font) SplitTextIntoLinesBySymbols(text string, size unit.PT, width unit.MM, textb []Text, hexb *[]unit.HEX) []Text {
	var (
		lineAdvance unit.EM
	)

	targetAdvance := width.PT().EM(size)
	hexbStartIndex := len(*hexb)

	for _, r := range text {
		_glyph := f.glyph(r)

		candidateAdvance := lineAdvance + _glyph.advance

		if candidateAdvance > targetAdvance {
			textb = append(textb, Text{
				content: (*hexb)[hexbStartIndex:],
				width:   lineAdvance.PT(size).MM(),
			})

			lineAdvance = 0
			hexbStartIndex = len(*hexb)
		}

		lineAdvance += _glyph.advance
		*hexb = append(*hexb, _glyph.index)
	}

	textb = append(textb, Text{
		content: (*hexb)[hexbStartIndex:],
		width:   lineAdvance.PT(size).MM(),
	})

	return textb
}

func (f *Font) SplitTextIntoLinesByWords(text string, size unit.PT, width unit.MM, textb []Text, hexb *[]unit.HEX) []Text {
	var (
		shift, wordAdvance, lineAdvance, spaceAdvance unit.EM
		wordsCount                                    int
		inWord                                        bool
	)

	targetAdvance := width.PT().EM(size)
	hexbStartIndex := len(*hexb)
	hexbEndIndex := hexbStartIndex
	wordStartIndex := hexbStartIndex

	for _, r := range text {
		_glyph := f.glyph(r)

		if !unicode.IsSpace(r) {
			if !inWord {
				inWord = true
				wordAdvance = 0
				wordStartIndex = len(*hexb)
				wordsCount++
			}

			wordAdvance += _glyph.advance
			*hexb = append(*hexb, _glyph.index)

			continue
		}

		if inWord {
			candidateAdvance := lineAdvance + spaceAdvance + wordAdvance

			if candidateAdvance > targetAdvance {
				if wordsCount > 1 {
					shift = (lineAdvance - targetAdvance) / unit.EM(wordsCount-1)
				}

				textb = append(textb, Text{
					content: (*hexb)[hexbStartIndex:hexbEndIndex],
					width:   lineAdvance.PT(size).MM(),
					shift:   shift,
				})

				hexbStartIndex = wordStartIndex
				hexbEndIndex = len(*hexb)
				lineAdvance = wordAdvance
				spaceAdvance = _glyph.advance
				wordsCount = 1
			} else {
				if r == lf {
					if wordsCount > 1 {
						shift = (candidateAdvance - targetAdvance) / unit.EM(wordsCount-1)
					}

					textb = append(textb, Text{
						content: (*hexb)[hexbStartIndex:],
						width:   candidateAdvance.PT(size).MM(),
						shift:   shift,
					})

					hexbStartIndex = len(*hexb)
					lineAdvance = 0
					spaceAdvance = 0
					wordsCount = 0
				} else {
					f.manager.saveSpaceGlyphIndex(_glyph.index)

					lineAdvance += wordAdvance + spaceAdvance
					spaceAdvance = _glyph.advance
					hexbEndIndex = len(*hexb)
					*hexb = append(*hexb, _glyph.index)
				}
			}

			inWord = false
		}
	}

	if inWord {
		candidateAdvance := lineAdvance + spaceAdvance + wordAdvance

		if candidateAdvance > targetAdvance {
			if wordsCount > 1 {
				shift = (lineAdvance - targetAdvance) / unit.EM(wordsCount-1)
			}

			textb = append(textb,
				Text{
					content: (*hexb)[hexbStartIndex:hexbEndIndex],
					width:   lineAdvance.PT(size).MM(),
					shift:   shift,
				},
				Text{
					content: (*hexb)[wordStartIndex:],
					width:   wordAdvance.PT(size).MM(),
					shift:   0,
				},
			)
		} else {
			if wordsCount > 1 {
				shift = (candidateAdvance - targetAdvance) / unit.EM(wordsCount-1)
			}

			textb = append(textb, Text{
				content: (*hexb)[hexbStartIndex:],
				width:   candidateAdvance.PT(size).MM(),
				shift:   shift,
			})
		}
	}

	return textb
}

func (f *Font) glyph(r rune) glyph {
	if r == lf {
		return glyph{}
	}

	f.manager.mu.RLock()
	defer f.manager.mu.RUnlock()

	if r < rusRunesLimitIndex {
		if f.manager.glyphsFastCache[r].rune == 0 {
			_glyph := f.saveGlyph(r)

			f.manager.glyphsFastCache[r] = _glyph
		}

		return f.manager.glyphsFastCache[r]
	}

	_glyph, ok := f.manager.glyphsSlowCache[r]
	if !ok {
		_glyph = f.saveGlyph(r)
	}

	return _glyph
}

func (f *Font) saveGlyph(r rune) glyph {
	gid, _ := f.face.GlyphIndex(f.manager.faceb, r) // err всегда nil

	advance, err := f.face.GlyphAdvance(f.manager.faceb, gid, ppem, hintingNone)
	if err != nil {
		advance = defaultAdvance // нас не интересует ошибка, просто ставим среднюю ширину символа.
	}

	_glyph := glyph{
		rune:    uint16(r),
		index:   unit.HEX(gid),
		advance: unit.Font(advance).EM(ppem),
	}

	f.manager.glyphsSlowCache[r] = _glyph
	f.manager.dirtyFlag = true

	return _glyph
}

func (f *Font) glyphs() []glyph {
	if !f.manager.dirtyFlag {
		return f.manager.glyphs
	}

	glyphs := slices.SortedFunc(maps.Values(f.manager.glyphsSlowCache), func(_glyph1 glyph, _glyph2 glyph) int {
		return cmp.Compare(_glyph1.index, _glyph2.index)
	})

	f.manager.glyphs = glyphs

	return glyphs
}

func (f *Font) glyphsWidthTable() parameter.GlyphsWidthTable {
	glyphs := f.glyphs()

	widthTable := make(parameter.GlyphsWidthTable, 0, len(glyphs))

	for _, _glyph := range glyphs {
		widthTable = append(widthTable,
			[2]uint64{
				uint64(_glyph.index),
				uint64(_glyph.advance.Font(ppem)),
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

type glyph struct {
	rune    uint16
	index   unit.HEX
	advance unit.EM
}

type fontManager struct {
	mu                 *sync.RWMutex
	compressor         *compressor.Compressor
	faceb              *sfnt.Buffer
	glyphs             []glyph
	glyphsFastCache    []glyph
	glyphsSlowCache    map[rune]glyph
	spaceGlyphsIndexes map[unit.HEX]struct{}
	dirtyFlag          bool
}

func (mgr *fontManager) glyph(r rune) (glyph, bool) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	if r < rusRunesLimitIndex {
		return mgr.glyphsFastCache[r], true
	}

	_glyph, ok := mgr.glyphsSlowCache[r]

	return _glyph, ok
}

func (mgr *fontManager) saveSpaceGlyphIndex(gid unit.HEX) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	mgr.spaceGlyphsIndexes[gid] = struct{}{}
}

func (mgr *fontManager) isSpaceGlyphIndex(gid unit.HEX) bool {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	_, ok := mgr.spaceGlyphsIndexes[gid]

	return ok
}

type Text struct {
	content primitives.String
	matrix  primitives.Matrix
	width   unit.MM
	shift   unit.EM
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
