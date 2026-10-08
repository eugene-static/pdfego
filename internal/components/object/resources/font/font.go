package font

import (
	"os"
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
	hintingNone = 0

	ppem           fixed.Int26_6 = 1000 << 6
	defaultAdvance fixed.Int26_6 = 600

	lf = '\n'
)

var DefaultRangeTable = []*unicode.RangeTable{
	unicode.Latin,
	unicode.Cyrillic,
	unicode.Greek,
}

var CommonRangeTable = []*unicode.RangeTable{
	unicode.White_Space,
	unicode.Punct,
	unicode.Number,
	unicode.Symbol,
}

type Fonts []*Font

func (f *Fonts) Add(font *Font) {
	font.id = uint8(len(*f))

	*f = append(*f, font)
}

func (f *Fonts) ForEach(fn func(*Font)) {
	for _, font := range *f {
		if len(font.manager.glyphsCache) > 0 {
			fn(font)
		}
	}
}

func (f *Fonts) Height(size unit.PT) unit.PT {
	height := size

	for _, font := range *f {
		h := font.Height(size)
		if h > height {
			height = h
		}
	}

	return height
}

func (f *Fonts) WriteToStream(dst *stream.Stream, text []Text, fontSize unit.PT) {
	if len(text) == 0 {
		return
	}

	bw := bytes.NewWriter(dst.AvailableBuffer())

	bw.Write(primitives.BeginText).LF()

	prevFontID := -1
	newLine := false

	for i := range text {
		newLine = true

		bw.
			Write(text[i].matrix).SP().
			Write(primitives.TextMatrix).LF()

		for j, s := range text[i].symbols {
			fontID := int(s.fontID)
			if fontID != prevFontID && fontID < len(*f) {
				prevFontID = fontID
				newLine = true

				if j > 0 {
					bw.
						WriteByte(primitives.HexadecimalStringClose).SP().
						Write(primitives.ShowText).LF()
				}

				bw.
					Write((*f)[s.fontID].Alias()).SP().
					Write(fontSize).SP().
					Write(primitives.TextFont).LF()
			}

			if newLine {
				newLine = false

				bw.WriteByte(primitives.HexadecimalStringOpen)
			}

			bw.Write(s.hex)
		}

		bw.
			WriteByte(primitives.HexadecimalStringClose).SP().
			Write(primitives.ShowText).LF()
	}

	bw.Write(primitives.EndText).LF()

	dst.Write(bw.Bytes())
}

func (f *Fonts) WriteToStreamWithIndividualGlyphPosition(dst *stream.Stream, text []Text, fontSize unit.PT) {
	bw := bytes.NewWriter(dst.AvailableBuffer())

	bw.Write(primitives.BeginText).LF()

	prevFontID := -1
	newLine := false

	for i := range text {
		newLine = true

		bw.
			Write(text[i].matrix).SP().
			Write(primitives.TextMatrix).LF()

		for j, s := range text[i].symbols {
			fontID := int(s.fontID)
			if fontID != prevFontID && fontID < len(*f) {
				prevFontID = fontID
				newLine = true

				if j > 0 {
					bw.
						WriteByte(primitives.HexadecimalStringClose).
						WriteByte(primitives.ArrayClose).SP().
						Write(primitives.ShowTextWithIndividualGlyphPositioning).LF()
				}

				bw.
					Write((*f)[s.fontID].Alias()).SP().
					Write(fontSize).SP().
					Write(primitives.TextFont).SP()
			}

			if newLine {
				newLine = false

				bw.
					WriteByte(primitives.ArrayOpen).
					WriteByte(primitives.HexadecimalStringOpen)
			}

			bw.Write(s.hex)

			if s.isSpace {
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

func (f *Fonts) TextToLineNotDivided(text string, size unit.PT, width unit.MM, textb []Text, symb *[]Symbol) []Text {
	var (
		shift       unit.EM
		lineAdvance unit.EM
		spaceCount  int
	)

	targetAdvance := width.PT().Sub(Padding(size) * 2).EM(size)
	symbStartIndex := len(*symb)

	for _, r := range text {
		_glyph, fontID := f.glyph(r)
		isSpace := false

		if unicode.IsSpace(r) {
			spaceCount++
			isSpace = true
		}

		lineAdvance += _glyph.advance
		*symb = append(*symb, Symbol{
			hex:     _glyph.index,
			fontID:  fontID,
			isSpace: isSpace,
		})
	}

	if spaceCount > 0 {
		shift = (lineAdvance - targetAdvance) / unit.EM(spaceCount)
	}

	textb = append(textb, Text{
		symbols: (*symb)[symbStartIndex:],
		width:   lineAdvance.PT(size).MM(),
		shift:   shift,
	})

	return textb
}

func (f *Fonts) TextToLineDividedBySymbols(text string, size unit.PT, width unit.MM, textb []Text, symb *[]Symbol) []Text {
	var (
		lineAdvance unit.EM
	)

	targetAdvance := width.PT().Sub(Padding(size) * 2).EM(size)
	symbStartIndex := len(*symb)

	for _, r := range text {
		_glyph, fontID := f.glyph(r)

		candidateAdvance := lineAdvance + _glyph.advance

		if candidateAdvance > targetAdvance {
			textb = append(textb, Text{
				symbols: (*symb)[symbStartIndex:],
				width:   lineAdvance.PT(size).MM(),
			})

			lineAdvance = 0
			symbStartIndex = len(*symb)
		}

		lineAdvance += _glyph.advance

		*symb = append(*symb, Symbol{
			hex:    _glyph.index,
			fontID: fontID,
		})
	}

	textb = append(textb, Text{
		symbols: (*symb)[symbStartIndex:],
		width:   lineAdvance.PT(size).MM(),
	})

	return textb
}

func (f *Fonts) TextToLineDividedByWords(text string, size unit.PT, width unit.MM, textb []Text, symb *[]Symbol) []Text {
	var (
		shift, wordAdvance, lineAdvance, spaceAdvance unit.EM
		wordsCount                                    int
		inWord                                        bool
	)

	targetAdvance := width.PT().Sub(Padding(size) * 2).EM(size)
	lineStartIndex := len(*symb)
	lineEndIndex := lineStartIndex
	wordStartIndex := lineStartIndex

	for _, r := range text {
		_glyph, fontID := f.glyph(r)

		if !unicode.IsSpace(r) {
			if !inWord {
				inWord = true
				wordAdvance = 0
				wordStartIndex = len(*symb)
				wordsCount++
			}

			wordAdvance += _glyph.advance

			*symb = append(*symb, Symbol{
				hex:    _glyph.index,
				fontID: fontID,
			})

			continue
		}

		if inWord {
			inWord = false

			candidateAdvance := lineAdvance + spaceAdvance + wordAdvance

			// строка не поместилась
			if candidateAdvance > targetAdvance {
				if wordsCount > 1 {
					shift = 0
					// мы на минимум третьем слове, значит в строке их два, а пробел будет один
					if wordsCount > 2 {
						shift = (lineAdvance - targetAdvance) / unit.EM(wordsCount-2)
					}

					textb = append(textb, Text{
						symbols: (*symb)[lineStartIndex:lineEndIndex],
						width:   lineAdvance.MM(size),
						shift:   shift,
					})
				}

				// если совпало так, что и строка не помещается, и мы на символе переноса, то надо записать непомещающееся слово отдельной строкой
				// или если первое же слово шире строки, мы его запишем, так как делить нечего
				if r == lf || wordsCount == 1 {
					textb = append(textb, Text{
						symbols: (*symb)[wordStartIndex:],
						width:   wordAdvance.MM(size),
					})

					lineStartIndex = len(*symb)
					lineEndIndex = lineStartIndex
					spaceAdvance = 0
					lineAdvance = 0
					wordsCount = 0

					continue
				}

				// мы записали строку, перенесли слово на следующей строку, пробел нужно добавить
				lineStartIndex = wordStartIndex
				lineEndIndex = len(*symb)
				lineAdvance = wordAdvance
				spaceAdvance = _glyph.advance
				wordsCount = 1

				*symb = append(*symb, Symbol{
					hex:     _glyph.index,
					fontID:  fontID,
					isSpace: true,
				})

				continue
			}

			// если попался символ переноса, то просто пишем всю строку
			if r == lf {
				shift = 0
				if wordsCount > 1 {
					shift = (candidateAdvance - targetAdvance) / unit.EM(wordsCount-1)
				}

				textb = append(textb, Text{
					symbols: (*symb)[lineStartIndex:],
					width:   candidateAdvance.MM(size),
					shift:   shift,
				})

				lineStartIndex = len(*symb)
				lineEndIndex = lineStartIndex
				spaceAdvance = 0
				lineAdvance = 0
				wordsCount = 0

				continue
			}

			// просто пишем пробел
			lineAdvance += wordAdvance + spaceAdvance
			spaceAdvance = _glyph.advance
			lineEndIndex = len(*symb)

			*symb = append(*symb, Symbol{
				hex:     _glyph.index,
				fontID:  fontID,
				isSpace: true,
			})
		}
	}

	switch wordsCount {
	case 0:
		return textb
	case 1:
		textb = append(textb, Text{
			symbols: (*symb)[wordStartIndex:],
			width:   wordAdvance.MM(size),
		})

		return textb
	default:
		candidateAdvance := lineAdvance + spaceAdvance + wordAdvance

		// если последнее слово не поместилось, надо записать две строки
		if candidateAdvance > targetAdvance {
			shift = 0
			if wordsCount > 2 {
				shift = (lineAdvance - targetAdvance) / unit.EM(wordsCount-2)
			}

			textb = append(textb,
				Text{
					symbols: (*symb)[lineStartIndex:lineEndIndex],
					width:   lineAdvance.MM(size),
					shift:   shift,
				},
				Text{
					symbols: (*symb)[wordStartIndex:],
					width:   wordAdvance.MM(size),
				},
			)
		} else {
			// если строка помещается, пишем строку целиком
			shift = (candidateAdvance - targetAdvance) / unit.EM(wordsCount-1)

			textb = append(textb, Text{
				symbols: (*symb)[lineStartIndex:],
				width:   candidateAdvance.MM(size),
				shift:   shift,
			})
		}
	}

	return textb
}

func (f *Fonts) glyph(r rune) (glyph, uint8) {
	if r == lf {
		return glyph{}, 0
	}

	for _, font := range *f {
		if unicode.IsOneOf(font.manager.rangeTables, r) {
			g, ok := font.manager.getGlyph(r)
			if !ok {
				g = font.saveGlyph(r)
			}

			return g, font.id
		}
	}

	return defaultGlyph(), 0
}

type Font struct {
	id         uint8
	alias      primitives.Alias
	face       *sfnt.Font
	faceb      *sfnt.Buffer
	compressor *compressor.Compressor
	bytes      []byte
	manager    manager
	metrics    metrics
	objects    objects
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

type Options struct {
	Alias       string
	Path        string
	Compressor  *compressor.Compressor
	RangeTables []*unicode.RangeTable
}

func New(options Options) (*Font, error) {
	fontBytes, err := os.ReadFile(options.Path)
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

	if len(options.RangeTables) == 0 {
		options.RangeTables = DefaultRangeTable
	}

	rangeTables := append(CommonRangeTable, options.RangeTables...)

	_manager := manager{
		mu:          &sync.RWMutex{},
		glyphsCache: make(map[unit.Rune]glyph),
		rangeTables: rangeTables,
	}

	f := &Font{
		alias:      primitives.Alias(options.Alias),
		face:       sfntFont,
		faceb:      buf,
		compressor: options.Compressor,
		bytes:      fontBytes,
		metrics:    _metrics,
		manager:    _manager,
	}

	return f, nil
}

func (f *Font) Alias() parameter.Name {
	return parameter.Name(f.alias) + parameter.Name(bytes.FormatUint(uint(f.id)))
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

func (f *Font) saveGlyph(r rune) glyph {
	gid, _ := f.face.GlyphIndex(f.faceb, r) // err всегда nil

	advance, err := f.face.GlyphAdvance(f.faceb, gid, ppem, hintingNone)
	if err != nil {
		advance = defaultAdvance // нас не интересует ошибка, просто ставим среднюю ширину символа.
	}

	_glyph := glyph{
		rune:    unit.Rune(r),
		index:   unit.HEX(gid),
		advance: unit.Font(advance).EM(ppem),
	}

	f.manager.saveGlyph(r, _glyph)

	return _glyph
}

func (f *Font) glyphsWidthTable() parameter.GlyphsWidthTable {
	glyphs := f.manager.sortedGlyphs()

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

	bytes, err := f.compressor.ForceCompress(subset)
	if err != nil {
		return err
	}

	str.Write(bytes)

	return nil
}

// Padding -- расстояние от текста до границ объекта, в котором он расположен.
func Padding(fontSize unit.PT) unit.PT {
	return fontSize / 7
}

func UnderlineSize(fontSize unit.PT) unit.PT {
	return fontSize / 30
}
