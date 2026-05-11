package template

import (
	"cmp"
	"maps"
	"os"
	"slices"
	"sync"

	"github.com/cdillond/gdf"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

const (
	hintingNone = 0
)

type font struct {
	alias      string
	face       *sfnt.Font
	rawData    []byte
	subsetData []byte
	manager    fontManager
	metrics    fontMetrics
}

type fontManager struct {
	mu          sync.RWMutex
	glyphsCache map[rune]*glyph
	glyphs      []*glyph
	dirtyFlag   bool
}

type glyph struct {
	rune    uint16
	index   sfnt.GlyphIndex
	advance fixed.Int26_6
}

type fontMetrics struct {
	ascent    int
	descent   int
	capHeight int
	stemV     int
	ppem      fixed.Int26_6
	fontBBox  []int
}

func defaultGlyph() *glyph {
	return &glyph{
		rune:    '□',
		index:   0,
		advance: 600,
	}
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

// /Flags 4
// /FontBBox [-203 -303 1050 910]
// /ItalicAngle 0
// /Ascent 905
// /Descent -212
// /CapHeight 688
// /StemV 569
func (core *Core) setFont(path, alias string) error {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	sfntFont, err := sfnt.Parse(fontBytes)
	if err != nil {
		return err
	}

	ppem := fixed.I(1000)

	var buf sfnt.Buffer

	sfntMetrics, err := sfntFont.Metrics(&buf, ppem, hintingNone)
	if err != nil {
		return err
	}

	ascent := sfntMetrics.Ascent.Round()
	descent := sfntMetrics.Descent.Round()
	capHeight := sfntMetrics.CapHeight.Round()

	bounds, err := sfntFont.Bounds(&buf, ppem, hintingNone)
	if err != nil {
		return err
	}

	// [Min.X, Min.Y, Max.X, Max.Y]
	// Min.Y и Max.Y поменяны местами нарочно ввиду того, что у PDF начало координат находится в левом нижнем углу.
	fontBBox := []int{
		bounds.Min.X.Round(),
		-bounds.Max.Y.Round(),
		bounds.Max.X.Round(),
		-bounds.Min.Y.Round(),
	}

	// Среднее значение. Обычно такое и остается.
	glyphAdvanceRounded := 80

	glyphIndex, err := sfntFont.GlyphIndex(&buf, 'I')
	if err != nil {
		return err
	}

	glyphAdvance, err := sfntFont.GlyphAdvance(&buf, glyphIndex, ppem, hintingNone)
	if err != nil {
		return err
	}

	glyphAdvanceRounded = glyphAdvance.Round()
	if glyphAdvanceRounded > 20 && glyphAdvanceRounded < 200 {
		glyphAdvanceRounded = glyphAdvanceRounded * 3 / 4
	}

	stemV := glyphAdvanceRounded

	//b := new(bytes.Buffer)
	//w := zlib.NewWriter(b)
	//defer w.Close()
	//
	//_, err = w.Write(fontBytes)
	//if err != nil {
	//	return err
	//}

	metrics := fontMetrics{
		ascent:    ascent,
		descent:   descent,
		capHeight: capHeight,
		stemV:     stemV,
		ppem:      ppem,
		fontBBox:  fontBBox,
	}

	core.fonts[alias] = &font{
		alias:   alias,
		face:    sfntFont,
		rawData: fontBytes,
		metrics: metrics,
		manager: fontManager{
			mu:          sync.RWMutex{},
			glyphsCache: make(map[rune]*glyph),
		},
	}

	return nil
}

func (f *font) measureText(fontSize int, text string) float64 {
	var (
		advance fixed.Int26_6
		buf     sfnt.Buffer
	)

	fontSizeEm := fixed.I(fontSize)
	ppemFont := f.metrics.ppem.Mul(fontSizeEm)
	prevGlyphIndex := sfnt.GlyphIndex(0)

	for _, c := range text {
		gl := f.glyph(c)

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

	width := float64((advance / fixed.Int26_6(f.metrics.ppem.Round())).Round())

	return mm(width)
}

func (f *font) saveRunes(text string) {
	for _, r := range text {
		_, ok := f.manager.glyphsCache[r]
		if ok {
			continue
		}

		buf := new(sfnt.Buffer)

		index, _ := f.face.GlyphIndex(buf, r) //err is always nil

		adv, err := f.face.GlyphAdvance(buf, index, f.metrics.ppem, hintingNone)
		if err != nil {
			adv = 600
		}

		f.manager.addGlyph(r, index, adv)

		f.manager.dirtyFlag = true
	}
}

func (f *font) glyph(r rune) *glyph {
	gl, ok := f.manager.glyphsCache[r]
	if !ok {
		gl = defaultGlyph()
	}

	return gl
}

func (f *font) glyphs() []*glyph {
	if !f.manager.dirtyFlag {
		return f.manager.glyphs
	}

	glyphs := slices.SortedFunc(maps.Values(f.manager.glyphsCache), func(g *glyph, g2 *glyph) int {
		return cmp.Compare(g.index, g2.index)
	})

	f.manager.dirtyFlag = false
	f.manager.glyphs = glyphs

	return glyphs
}

func (f *font) subset() {
	if !f.manager.dirtyFlag {
		return
	}

	runes := make(map[rune]struct{})

	for r := range f.manager.glyphsCache {
		runes[r] = struct{}{}
	}

	subsetter := gdf.DefaultSubsetter{}

	subsetter.Init(f.face, f.rawData, "")

	fontTable, err := subsetter.Subset(runes)
	if err != nil {
		panic(err) //TODO:
	}

	sfntFont, err := sfnt.Parse(fontTable)
	if err != nil {
		panic(err)
	}

	f.face = sfntFont

	buf := new(sfnt.Buffer)

	for r, gl := range f.manager.glyphsCache {
		index, _ := sfntFont.GlyphIndex(buf, r)

		gl.index = index
	}

	f.subsetData = fontTable
}

func (mgr *fontManager) addGlyph(r rune, index sfnt.GlyphIndex, advance fixed.Int26_6) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	mgr.glyphsCache[r] = &glyph{
		rune:    uint16(r),
		index:   index,
		advance: advance,
	}
}

func (core *Core) getFont(alias string) *font {
	return core.fonts[alias]
}

func (core *Core) fontRegular() *font {
	return core.fonts[FontRegular]
}

func (core *Core) fontBold() *font {
	return core.fonts[FontBold]
}
