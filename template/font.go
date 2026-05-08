package template

import (
	"bytes"
	"cmp"
	"compress/zlib"
	"os"
	"slices"

	image_font "golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

type font struct {
	alias           string
	face            *sfnt.Font
	compressedData  *bytes.Buffer
	uncompressedLen int
	glyphMap        map[rune]struct{}
	glyphs          []glyph
	ascent          int
	descent         int
	capHeight       int
	stemV           int
	hinting         image_font.Hinting
	unitsPerEm      int
	fontBBox        []int
}

type glyph struct {
	rune    uint16
	index   uint16
	advance int
}

func (core *Core) SetFont(path, alias string) error {
	err := core.setFont(path, alias)
	if err != nil {
		return err
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

func (core *Core) SetDefaultFontSize(fontSize int) {
	core.fontSize = fontSize
	core.fontHeight = mm(float64(fontSize)) * 1.2
}

func (core *Core) setFont(path, alias string) error {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	sfntFont, err := sfnt.Parse(fontBytes)
	if err != nil {
		return err
	}

	unitsPerEm := int(sfntFont.UnitsPerEm())
	ppem := fixed.I(unitsPerEm)
	hinting := image_font.HintingNone

	sc := newScaler(unitsPerEm)

	var buf sfnt.Buffer

	metrics, err := sfntFont.Metrics(&buf, ppem, hinting)
	if err != nil {
		return err
	}

	ascent := sc.scale(metrics.Ascent.Round())
	descent := sc.scale(metrics.Descent.Round())
	capHeight := sc.scale(metrics.CapHeight.Round())

	bounds, err := sfntFont.Bounds(&buf, ppem, hinting)
	if err != nil {
		return err
	}

	fontBox := []int{
		sc.scale(bounds.Min.X.Round()),
		-sc.scale(bounds.Max.Y.Round()),
		sc.scale(bounds.Max.X.Round()),
		-sc.scale(bounds.Min.Y.Round()),
	}

	glyphAdvanceRounded := 80

	glyphIndex, err := sfntFont.GlyphIndex(&buf, 'I')
	if err != nil {
		return err
	}

	glyphAdvance, err := sfntFont.GlyphAdvance(&buf, glyphIndex, ppem, hinting)
	if err != nil {
		return err
	}

	glyphAdvanceRounded = glyphAdvance.Round()
	if glyphAdvanceRounded > 20 && glyphAdvanceRounded < 200 {
		glyphAdvanceRounded = glyphAdvanceRounded * 3 / 4
	}

	stemV := glyphAdvanceRounded

	b := new(bytes.Buffer)
	w := zlib.NewWriter(b)
	defer w.Close()

	_, err = w.Write(fontBytes)
	if err != nil {
		return err
	}

	core.fonts[alias] = &font{
		alias:           alias,
		face:            sfntFont,
		compressedData:  b,
		uncompressedLen: len(fontBytes),
		glyphMap:        make(map[rune]struct{}),
		glyphs:          make([]glyph, 0, 256),
		ascent:          ascent,
		descent:         -descent,
		capHeight:       capHeight,
		stemV:           stemV,
		fontBBox:        fontBox,
		unitsPerEm:      unitsPerEm,
		hinting:         hinting,
	}

	return nil
}

func (f *font) measureText(fontSize int, text string) float64 {
	var (
		advance fixed.Int26_6
		buf     sfnt.Buffer
	)

	ppemFont := fixed.I(fontSize)
	//ppem := fixed.I(f.unitsPerEm)
	prevGlyphIndex := sfnt.GlyphIndex(0)

	for _, c := range text {
		glyphIndex, _ := f.face.GlyphIndex(&buf, c) // err is always nil

		adv, err := f.face.GlyphAdvance(&buf, glyphIndex, ppemFont, f.hinting)
		if err != nil {
			//TODO: обработка ошибок
		}

		advance += adv

		if prevGlyphIndex > 0 {
			kern, err := f.face.Kern(&buf, prevGlyphIndex, glyphIndex, ppemFont, f.hinting)
			if err != nil {
				//TODO: обработка ошибок
			}

			advance += kern
		}

		prevGlyphIndex = glyphIndex
	}

	width := float64(advance >> 6)

	return mm(width)
}

func (f *font) saveIndex(buf *sfnt.Buffer, r rune, index sfnt.GlyphIndex) {
	_, ok := f.glyphMap[r]
	if ok {
		return
	}

	sc := newScaler(f.unitsPerEm)
	ppem := fixed.I(f.unitsPerEm)

	adv, err := f.face.GlyphAdvance(buf, index, ppem, f.hinting)
	if err != nil {
		adv = 600
	}

	f.glyphs = append(f.glyphs, glyph{
		rune:    uint16(r),
		index:   uint16(index),
		advance: sc.scale(adv.Round()),
	})

	f.glyphMap[r] = struct{}{}
}

func (f *font) glyphAdvances() []glyph {
	slices.SortFunc(f.glyphs, func(a, b glyph) int {
		return cmp.Compare(a.index, b.index)
	})

	return f.glyphs
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
