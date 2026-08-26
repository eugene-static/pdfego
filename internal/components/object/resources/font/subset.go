package font

import (
	"bytes"
	"encoding/binary"
	"errors"
	"slices"

	"github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
)

type tableTag uint32

const (
	cmap tableTag = 'c'<<24 | 'm'<<16 | 'a'<<8 | 'p'
	glyf tableTag = 'g'<<24 | 'l'<<16 | 'y'<<8 | 'f'
	head tableTag = 'h'<<24 | 'e'<<16 | 'a'<<8 | 'd'
	hhea tableTag = 'h'<<24 | 'h'<<16 | 'e'<<8 | 'a'
	hmtx tableTag = 'h'<<24 | 'm'<<16 | 't'<<8 | 'x'
	loca tableTag = 'l'<<24 | 'o'<<16 | 'c'<<8 | 'a'
	maxp tableTag = 'm'<<24 | 'a'<<16 | 'x'<<8 | 'p'
)

var pdfTables = [...]tableTag{
	//1330851634, // OS/2
	cmap, // cmap
	glyf, // glyf
	head, // head
	hhea, // hhea
	hmtx, // hmtx
	loca, // loca
	maxp, // maxp
	//1851878757, // name
	//1886352244, // post
}

// ttfSubset -- упрощенный сабсеттинг шрифта. Здесь не переписываются таблица, а "обнуляются" контуры неиспользуемых глифов.
// При таком способе индексы глифов не переписываются, поэтому можно использовать те, что были собраны до этого.
func (f *Font) ttfSubset() ([]byte, error) {
	glyphsIndexes := make([]uint32, 0, 256)
	glyphsIndexes = append(glyphsIndexes, notdef)

	glyphset := make(map[uint32]struct{}, len(glyphsIndexes))
	glyphset[0] = struct{}{}

	for _, gl := range f.glyphs() {
		glyphsIndexes = append(glyphsIndexes, uint32(gl.index))
		glyphset[uint32(gl.index)] = struct{}{}
	}

	srcR := bytes.NewReader(f.bytes)

	ld, err := opentype.NewLoader(srcR)
	if err != nil {
		return nil, err
	}

	headRaw, err := ld.RawTable(opentype.Tag(head))
	if err != nil {
		return nil, err
	}

	_head, _, err := tables.ParseHead(headRaw)
	if err != nil {
		return nil, err
	}

	isLong := _head.IndexToLocFormat == 1

	locaRaw, err := ld.RawTable(opentype.Tag(loca))
	if err != nil {
		return nil, err
	}

	_loca, err := tables.ParseLoca(locaRaw, f.face.NumGlyphs(), isLong)
	if err != nil {
		return nil, err
	}

	glyfRaw, err := ld.RawTable(opentype.Tag(glyf))
	if err != nil {
		return nil, err
	}

	// здесь включаются глифы, которые являются компонентами композитных глифов
	var composites []uint32

	n := uint32(len(_loca))

	for i := range glyphset {
		var offset, next uint32

		switch {
		case i < n-1:
			offset = _loca[i]
			next = _loca[i+1]
		case i == n-1:
			continue
		default:
			return nil, errors.New("неверный глиф-индекса или таблица Loca")
		}

		if next == offset {
			continue
		}

		// следуя спецификации, _loca[n] должен быть меньше или равен _loca[n+1]
		if next < offset || int(next) > len(glyfRaw) {
			return nil, errors.New("неверная таблица Loca")
		}

		g, _, err := tables.ParseGlyph(glyfRaw[offset:next])
		if err != nil {
			return nil, err
		}

		cGlyph, ok := g.Data.(tables.CompositeGlyph)
		if !ok {
			continue
		}

		for _, cGlyphPart := range cGlyph.Glyphs {
			_, seen := glyphset[uint32(cGlyphPart.GlyphIndex)]
			if !seen {
				composites = append(composites, uint32(cGlyphPart.GlyphIndex))
			}
		}
	}

	for _, comp := range composites {
		glyphset[comp] = struct{}{}
	}

	glyphsIndexes = append(glyphsIndexes, composites...)

	slices.Sort(glyphsIndexes)

	// loop back over the loca table and zero out the outlines of unused glyphs
	for i := 0; i < len(_loca); i++ {
		var offset, next uint32
		if i < len(_loca)-1 {
			offset = _loca[i]
			next = _loca[i+1]
		} else {
			offset = _loca[i]
			next = offset
		}

		if next < offset || int(next) > len(glyfRaw) {
			return nil, errors.New("неверная таблица Loca")
		}

		_, used := glyphset[uint32(i)]
		if !used {
			// zero out old glyph outlines
			for j := offset; j < next; j++ {
				glyfRaw[j] = 0
			}
		}
	}

	final := glyphsIndexes[len(glyphsIndexes)-1]
	// the loca table needs to be no more than final GID long
	finalOffset := _loca[final+1]

	// update the number of glyphs in the maxp table
	// https://learn.microsoft.com/en-us/typography/opentype/spec/maxp
	_maxp, err := ld.RawTable(opentype.Tag(maxp))
	if err != nil {
		return nil, err
	}

	if len(_maxp) >= 6 {
		// we can proceed
		binary.BigEndian.PutUint16(_maxp[4:], uint16(final+1))
	}

	// truncate the loca table
	// https://learn.microsoft.com/en-us/typography/opentype/spec/loca
	// **"In order to compute the length of the last glyph element, there is an extra entry after the last valid index."**
	// The resulting slice includes the glyph location data that begins at the final valid offset as well as the data
	// that begins at the offset after that.
	if isLong {
		// each offset is 4 bytes
		locaRaw = locaRaw[:4*(final+1)+4+1]
	} else {
		// each offset is 2 bytes
		locaRaw = locaRaw[:2*(final+1)+2+1]
	}

	// truncate the glyf table
	glyfRaw = glyfRaw[:finalOffset]

	tables := make([]opentype.Table, len(pdfTables))
	for i, tag := range pdfTables {
		switch tag {
		case glyf:
			tables[i] = opentype.Table{Content: glyfRaw, Tag: opentype.Tag(tag)}
		case head:
			tables[i] = opentype.Table{Content: headRaw, Tag: opentype.Tag(tag)}
		case loca:
			tables[i] = opentype.Table{Content: locaRaw, Tag: opentype.Tag(tag)}
		case maxp:
			tables[i] = opentype.Table{Content: _maxp, Tag: opentype.Tag(tag)}
		default:
			cnt, err := ld.RawTable(opentype.Tag(tag))
			if err != nil {
				return nil, err
			}

			tables[i] = opentype.Table{Content: cnt, Tag: opentype.Tag(tag)}
		}
	}

	return opentype.WriteTTF(tables), nil
}
