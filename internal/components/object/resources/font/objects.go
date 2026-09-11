package font

import (
	"slices"

	"github.com/eugene-static/pdfego/internal/components/dictionary"
	"github.com/eugene-static/pdfego/internal/components/object"
	"github.com/eugene-static/pdfego/internal/components/parameter"
	"github.com/eugene-static/pdfego/internal/components/stream/bytes"
	"github.com/eugene-static/pdfego/internal/components/stream/primitives"
	"github.com/eugene-static/pdfego/unit"
)

type objects struct {
	cMapB          *object.Object
	fontFile2      *object.Object
	fontDescriptor *object.Object
	cidFontType2   *object.Object
	fOntType0      *object.Object
}

type cMapB struct {
	Length parameter.Integer `pdf:"Length"`
}

func (f *Font) CMapB() *object.Object {
	f.manager.mu.Lock()
	defer f.manager.mu.Unlock()

	if f.objects.cMapB == nil {
		f.objects.cMapB = object.New()
	}

	if f.manager.dirtyFlag {
		glyphs := f.glyphs()
		stream := f.objects.cMapB.Stream()

		stream.Reset()

		stream.WriteString("/CIDInit /ProcSet findresource begin\n12 dict begin\nbegincmap\n/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n/CMapName /Adobe-Identity-UCS def\n/CMapType 2 def\n1 begincodespacerange\n<0000> <FFFF>\nendcodespacerange\n")

		bw := bytes.NewWriter(stream.AvailableBuffer())

		for chunk := range slices.Chunk(glyphs, 100) {
			bw = bw.
				WriteInt(len(chunk)).SP().
				Write(primitives.BeginBfChar).LF()

			for _, _glyph := range chunk {
				// <0000> <FFFF>
				bw = bw.
					WriteByte(primitives.HexadecimalStringOpen).
					Write(_glyph.index).
					WriteByte(primitives.HexadecimalStringClose).
					SP().
					WriteByte(primitives.HexadecimalStringOpen).
					Write(unit.HEX(_glyph.rune)).
					WriteByte(primitives.HexadecimalStringClose).
					LF()
			}

			bw = bw.
				Write(primitives.EndBfChar).
				LF()
		}

		stream.Write(bw.Bytes())
		stream.WriteString("endcmap\n/CMapName currentdict /CMap defineresource pop\nend\nend")

		_cMapB := &cMapB{
			Length: parameter.Integer(stream.Len()),
		}

		f.objects.cMapB.ReadDictionary(_cMapB)
	}

	return f.objects.cMapB
}

type fontFile2 struct {
	Length  parameter.Integer `pdf:"Length"`
	Length1 parameter.Integer `pdf:"Length1"`
	Filter  parameter.Name    `pdf:"Filter"`
}

func (f *Font) FontFile2() (*object.Object, error) {
	f.manager.mu.Lock()
	defer f.manager.mu.Unlock()

	if f.objects.fontFile2 == nil {
		f.objects.fontFile2 = object.New()
	}

	if f.manager.dirtyFlag {
		stream := f.objects.fontFile2.Stream()

		err := f.subset(stream)
		if err != nil {
			return nil, err
		}

		_fontFile2 := &fontFile2{
			Length:  parameter.Integer(stream.Len()),
			Length1: parameter.Integer(len(f.bytes)),
			Filter:  parameter.FlateDecode,
		}

		f.objects.fontFile2.ReadDictionary(_fontFile2)
	}

	return f.objects.fontFile2, nil
}

type fontDescriptor struct {
	Type        parameter.Name       `pdf:"Type"`
	FontName    parameter.Name       `pdf:"FontName"`
	Flags       parameter.Integer    `pdf:"Flags"`
	FontBBox    parameter.Numbers    `pdf:"FontBBox"`
	ItalicAngle parameter.Number     `pdf:"ItalicAngle"`
	Ascent      parameter.Integer    `pdf:"Ascent"`
	Descent     parameter.Integer    `pdf:"Descent"`
	CapHeight   parameter.Integer    `pdf:"CapHeight"`
	StemV       parameter.Integer    `pdf:"StemV"`
	FontFile2   *parameter.Reference `pdf:"FontFile2"`
}

func (f *Font) FontDescriptor() *object.Object {
	f.manager.mu.Lock()
	defer f.manager.mu.Unlock()

	if f.objects.fontDescriptor == nil {
		f.objects.fontDescriptor = object.NewSimple()

		fontFile2Ref := new(parameter.Reference)

		if f.objects.fontFile2 != nil {
			fontFile2Ref = f.objects.fontFile2.Reference()
		}

		_fontDescriptor := &fontDescriptor{
			Type:        "FontDescriptor",
			FontName:    parameter.Name(f.alias),
			Flags:       4,
			FontBBox:    parameter.NewNumberArray(f.metrics.fontBBox...),
			ItalicAngle: parameter.Number(f.metrics.italicAngle),
			Ascent:      parameter.Integer(f.metrics.ascent),
			Descent:     parameter.Integer(f.metrics.descent),
			CapHeight:   parameter.Integer(f.metrics.capHeight.Font(ppem)),
			StemV:       parameter.Integer(f.metrics.stemV),
			FontFile2:   fontFile2Ref,
		}

		f.objects.fontDescriptor.ReadDictionary(_fontDescriptor)
	}

	return f.objects.fontDescriptor
}

type cidFontType2 struct {
	Type           parameter.Name             `pdf:"Type"`
	Subtype        parameter.Name             `pdf:"Subtype"`
	BaseFont       parameter.Name             `pdf:"BaseFont"`
	CIDSystemInfo  dictionary.Dictionary      `pdf:"CIDSystemInfo"`
	FontDescriptor *parameter.Reference       `pdf:"FontDescriptor"`
	DW             parameter.Integer          `pdf:"DW"`
	W              parameter.GlyphsWidthTable `pdf:"W"`
}

func (f *Font) CIDFontType2() *object.Object {
	f.manager.mu.Lock()
	defer f.manager.mu.Unlock()

	if f.objects.cidFontType2 == nil {
		f.objects.cidFontType2 = object.NewSimple()
	}

	if f.manager.dirtyFlag {
		fontDescriptorRef := new(parameter.Reference)

		if f.objects.fontDescriptor != nil {
			fontDescriptorRef = f.objects.fontDescriptor.Reference()
		}

		cidSystemInfo := *dictionary.New(
			parameter.New("Registry", parameter.Text("Adobe")),
			parameter.New("Ordering", parameter.Text("Identity")),
			parameter.New("Supplement", parameter.Integer(0)),
		)

		_cidFontType2 := &cidFontType2{
			Type:           "Font",
			Subtype:        "CIDFontType2",
			BaseFont:       parameter.Name(f.alias),
			CIDSystemInfo:  cidSystemInfo,
			FontDescriptor: fontDescriptorRef,
			DW:             600,
			W:              f.glyphsWidthTable(),
		}

		f.objects.cidFontType2.ReadDictionary(_cidFontType2)
	}

	return f.objects.cidFontType2
}

type fontType0 struct {
	Type            parameter.Name           `pdf:"Type"`
	Subtype         parameter.Name           `pdf:"Subtype"`
	BaseFont        parameter.Name           `pdf:"BaseFont"`
	Encoding        parameter.Name           `pdf:"Encoding"`
	DescendantFonts parameter.ReferencesPtrs `pdf:"DescendantFonts"`
	ToUnicode       *parameter.Reference     `pdf:"ToUnicode"`
}

func (f *Font) FontType0() *object.Object {
	f.manager.mu.Lock()
	defer f.manager.mu.Unlock()

	if f.objects.fOntType0 == nil {
		f.objects.fOntType0 = object.NewSimple()

		cidFontType2Ref := new(parameter.Reference)

		if f.objects.cidFontType2 != nil {
			cidFontType2Ref = f.objects.cidFontType2.Reference()
		}

		cMapBRef := new(parameter.Reference)

		if f.objects.cMapB != nil {
			cMapBRef = f.objects.cMapB.Reference()
		}

		_fontType0 := &fontType0{
			Type:            "Font",
			Subtype:         "Type0",
			BaseFont:        parameter.Name(f.alias),
			Encoding:        parameter.IdentityH,
			DescendantFonts: parameter.ReferencesPtrs{cidFontType2Ref},
			ToUnicode:       cMapBRef,
		}

		f.objects.fOntType0.ReadDictionary(_fontType0)
	}

	f.manager.dirtyFlag = false

	return f.objects.fOntType0
}
