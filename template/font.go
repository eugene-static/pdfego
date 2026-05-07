package template

import (
	"bytes"
	"os"

	font_face "golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

type font struct {
	alias      string
	face       *sfnt.Font
	buf        *bytes.Buffer
	ascent     int
	descent    int
	capHeight  int
	stemV      int
	fontBBox   []int64
	unitsPerEm int32
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

func (core *Core) SetDefaultFontSize(fontSize float64) {
	core.fontSize = fontSize
	core.fontHeight = mm(fontSize) * 1.2
}

func (core *Core) setFont(alias, path string) error {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	sfntFont, err := sfnt.Parse(fontBytes)
	if err != nil {
		return err
	}

	unitsPerEm := int32(sfntFont.UnitsPerEm())

	scale := fixed.Int26_6(unitsPerEm)

	var buf sfnt.Buffer

	metrics, err := sfntFont.Metrics(&buf, scale, font_face.HintingNone)
	if err != nil {
		return err
	}

	ascent := metrics.Ascent.Round()
	descent := metrics.Descent.Round()
	capHeight := metrics.CapHeight.Round()

	bounds, err := sfntFont.Bounds(&buf, scale, font_face.HintingNone)
	if err != nil {
		return err
	}

	fontBox := []int64{
		int64(bounds.Min.X.Round()),
		int64(bounds.Min.Y.Round()),
		int64(bounds.Max.X.Round()),
		int64(bounds.Max.Y.Round()),
	}

	stemV := 80

	glyphIndex, err := sfntFont.GlyphIndex(&buf, 'I')
	if err != nil {
		return err
	}

	glyphAdvance, err := sfntFont.GlyphAdvance(&buf, glyphIndex, scale, font_face.HintingNone)
	if err != nil {
		return err
	}

	glyphAdvanceRounded := glyphAdvance.Round()
	if glyphAdvanceRounded > 20 && glyphAdvanceRounded < 200 {
		stemV = glyphAdvanceRounded * 3 / 4
	}

	core.fonts[alias] = &font{
		alias:      alias,
		face:       sfntFont,
		buf:        bytes.NewBuffer(fontBytes),
		ascent:     ascent,
		descent:    descent,
		capHeight:  capHeight,
		stemV:      stemV,
		fontBBox:   fontBox,
		unitsPerEm: unitsPerEm,
	}

	return nil
}

func (core *Core) measureText(fontAlias string, fontSize float64, text string) float64 {
	f := core.getFont(fontAlias)

	const hinting = 0

	var (
		advance fixed.Int26_6
		buf     sfnt.Buffer
	)

	scale := fixed.Int26_6(fontSize * 64)
	prevGlyphIndex := sfnt.GlyphIndex(0)

	for _, c := range text {
		glyphIndex, _ := f.face.GlyphIndex(&buf, c) // err is always nil

		adv, err := f.face.GlyphAdvance(&buf, glyphIndex, scale, hinting)
		if err != nil {
			//TODO: обработка ошибок
		}

		advance += adv

		if prevGlyphIndex > 0 {
			kern, err := f.face.Kern(&buf, prevGlyphIndex, glyphIndex, scale, hinting)
			if err != nil {
				//TODO: обработка ошибок
			}

			advance += kern
		}

		prevGlyphIndex = glyphIndex
	}

	width := float64(advance) / 64.0

	return mm(width)
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
