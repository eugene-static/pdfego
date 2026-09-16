package font

import (
	"cmp"
	"maps"
	"slices"
	"sync"
	"unicode"

	"github.com/eugene-static/pdfego/unit"
)

type manager struct {
	mu          *sync.RWMutex
	glyphs      []glyph
	rangeTables []*unicode.RangeTable
	glyphsCache map[unit.Rune]glyph
	dirtyFlag   bool
}

type glyph struct {
	fontID  int
	rune    unit.Rune
	index   unit.HEX
	advance unit.EM
}

func defaultGlyph() glyph {
	return glyph{
		fontID:  0,
		rune:    0,
		index:   0,
		advance: unit.Font(defaultAdvance).EM(ppem),
	}
}

func (mgr *manager) getGlyph(r rune) (glyph, bool) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	_glyph, ok := mgr.glyphsCache[unit.Rune(r)]

	return _glyph, ok
}

func (mgr *manager) saveGlyph(r rune, glyph glyph) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	mgr.glyphsCache[unit.Rune(r)] = glyph
	mgr.dirtyFlag = true
}

func (mgr *manager) sortedGlyphs() []glyph {
	if !mgr.dirtyFlag {
		return mgr.glyphs
	}

	glyphs := slices.SortedFunc(maps.Values(mgr.glyphsCache), func(_glyph1 glyph, _glyph2 glyph) int {
		return cmp.Compare(_glyph1.index, _glyph2.index)
	})

	mgr.glyphs = glyphs

	return glyphs
}
