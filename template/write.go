package template

import "time"

const (
	pages     = 1
	resources = 2
)

func (core *Core) fillBuffer() {
	core.writeFileHeader()
	core.writePages()
	core.writeResources()

	infoObj := core.writeInfo()
	rootObj := core.writeCatalog()
	xrefOffset := core.writeXref()

	core.writeTrailer(rootObj, infoObj)
	core.writeEOF(xrefOffset)
}

func (core *Core) writeResources() {
	type fontResource struct {
		alias  string
		objNum int64
	}

	fontResources := make([]fontResource, 0, len(core.fonts))

	for _, f := range core.fonts {
		fontObjNum := core.writeFont(f)

		fontResources = append(fontResources, fontResource{
			alias:  f.alias,
			objNum: fontObjNum,
		})
	}

	b := core.mainBuffer

	core.setOffset(resources)

	b.startObj(resources)
	b.openObjectParameters()
	b.printFieldString("/Font", "")
	b.openObjectParameters()

	for _, resource := range fontResources {
		b.printRef("/"+resource.alias, resource.objNum)
	}

	b.closeObjectParameters()
	b.closeObjectParameters()
	b.endObj()
}

func (core *Core) writeFont(f *font) int64 {
	b := core.mainBuffer
	alias := "/" + f.alias

	// ---------- (Type0) ----------
	// "9 0 obj<< /Type /Font /Subtype /Type0 /BaseFont /%font_name /Encoding /Identity-H /DescendantFonts [10 0 R] >>endobj\n"
	fontNum := core.getObjNum()

	core.appendOffset()

	b.startObj(fontNum)
	b.openObjectParameters()
	b.printFieldString("/Type", "/Font")
	b.printFieldString("/Subtype", "/Type0")
	b.printFieldString("/BaseFont", alias)
	b.printFieldString("/Encoding", "/Identity-H")
	b.printRefArray("/DescendantFonts", []int64{fontNum + 1})
	b.closeObjectParameters()
	b.endObj()

	// ---------- (CIDFontType2) ----------
	// "10 0 obj<< /Type /Font /Subtype /CIDFontType2 /BaseFont /%font_name /CIDSystemInfo<< /Registry(Adobe) /Ordering(Identity) /Supplement 0 >> /FontDescriptor 11 0 R /DW 1000 >>endobj\n"
	objNum := core.getObjNum()

	core.appendOffset()

	b.startObj(objNum)
	b.openObjectParameters()
	b.printFieldString("/Type", "/Font")
	b.printFieldString("/Subtype", "/CIDFontType2")
	b.printFieldString("/BaseFont", alias)
	b.printFieldString("/CIDSystemInfo", "<< /Registry(Adobe) /Ordering(Identity) /Supplement 0 >>")
	b.printRef("/FontDescriptor", objNum+1)
	b.printFieldInt("/DW", 600)
	b.printGlyphWidthTable(f.glyphAdvances())
	b.closeObjectParameters()
	b.endObj()

	// ---------- (FontDescriptor) ----------
	// "11 0 obj<< /Type /FontDescriptor /FontName /%font_name /Flags 4 /FontBBox[-50 -200 800 800] /ItalicAngle 0 /Ascent 800 /Descent -200 /CapHeight 700 /StemV 80 /FontFile2 12 0 R >>endobj\n"
	objNum = core.getObjNum()

	core.appendOffset()

	b.startObj(objNum)
	b.openObjectParameters()
	b.printFieldString("/Type", "/FontDescriptor")
	b.printFieldString("/FontName", alias)
	b.printFieldInt("/Flags", 4)
	b.printFieldIntArray("/FontBBox", f.fontBBox)
	b.printFieldInt("/ItalicAngle", 0) //TODO: Italic Font
	b.printFieldInt("/Ascent", f.ascent)
	b.printFieldInt("/Descent", f.descent)
	b.printFieldInt("/CapHeight", f.capHeight)
	b.printFieldInt("/StemV", f.stemV)
	b.printRef("/FontFile2", objNum+1)
	b.closeObjectParameters()
	b.endObj()

	// ---------- (FontFile2) ----------
	// "12 0 obj<< /Length %font_bytes_length /Length1 %font_bytes_length >>stream\nfont_bytes\nendstream\nendobj\n"
	objNum = core.getObjNum()

	core.appendOffset()

	b.startObj(objNum)
	b.openObjectParameters()
	b.printFieldString("/Filter", "/FlateDecode")
	b.printFieldInt("/Length", f.compressedData.Len())
	b.printFieldInt("/Length1", f.uncompressedLen)
	b.closeObjectParameters()
	b.startStream()
	b.writeFrom(f.compressedData)
	b.endStream()
	b.endObj()

	return fontNum
}

func (core *Core) writePages() {
	pageObjs := make([]int64, 0, len(core.pageBuffers))

	for _, buf := range core.pageBuffers {
		pageObj := core.writePage(buf)

		pageObjs = append(pageObjs, pageObj)
	}

	b := core.mainBuffer

	core.setOffset(pages)

	b.startObj(pages)
	b.openObjectParameters()
	b.printFieldString("/Type", "/Pages")
	b.printRefArray("/Kids", pageObjs)
	b.printFieldInt("/Count", len(pageObjs))
	b.printFieldFloatArray("/MediaBox", []float64{0, 0, core.page.width, core.page.height})
	b.closeObjectParameters()
	b.endObj()
}

func (core *Core) writePage(buf *buffer) int64 {
	b := core.mainBuffer
	pageObj := core.getObjNum()

	core.appendOffset()

	b.startObj(pageObj)
	b.openObjectParameters()
	b.printFieldString("/Type", "/Page")
	b.printRef("/Parent", 1)
	b.printRef("/Resources", 2)
	b.printRef("/Contents", pageObj+1)
	b.closeObjectParameters()
	b.endObj()

	objNum := core.getObjNum()

	core.appendOffset()

	b.startObj(objNum)
	b.openObjectParameters()
	b.printFieldInt("/Length", buf.content.Len())
	b.closeObjectParameters()
	b.startStream()
	b.writeFrom(buf.content)
	b.endStream()
	b.endObj()

	return pageObj
}

func (core *Core) writeFileHeader() {
	core.mainBuffer.print("%PDF-1.4\n")
	core.mainBuffer.print("%\x80\x80\x80\x80\n")
}

func (core *Core) writeInfo() int64 {
	creationDate := time.Now().Format("D:20060102150405-07'00'")

	b := core.mainBuffer
	objNum := core.getObjNum()

	core.appendOffset()

	b.startObj(objNum)
	b.openObjectParameters()
	b.printFieldStringWithBrackets("/Producer", "OVP")
	b.printFieldStringWithBrackets("/CreationDate", creationDate)
	b.closeObjectParameters()
	b.endObj()

	return objNum
}

func (core *Core) writeCatalog() int64 {
	b := core.mainBuffer
	objNum := core.getObjNum()

	core.appendOffset()

	b.startObj(objNum)
	b.openObjectParameters()
	b.printFieldString("/Type", "/Catalog")
	b.printRef("/Pages", pages)
	b.closeObjectParameters()
	b.endObj()

	return objNum
}

func (core *Core) writeXref() int {
	xrefOffset := core.mainBuffer.content.Len()

	b := core.mainBuffer

	b.print("xref\n")
	b.printFieldInt("0", len(core.offsets))
	b.print("0000000000 65535 f\n")
	for i := range core.offsets {
		if i == 0 {
			continue
		}

		b.printXref(core.offsets[i])
	}

	return xrefOffset
}

func (core *Core) writeTrailer(root, info int64) {
	b := core.mainBuffer

	b.print("trailer\n")
	b.openObjectParameters()
	b.printFieldInt("/Size", len(core.offsets))
	b.printRef("/Root", root)
	b.printRef("/Info", info)
	b.closeObjectParameters()
}

func (core *Core) writeEOF(xrefOffset int) {
	b := core.mainBuffer

	b.print("startxref\n")
	b.printInt64(int64(xrefOffset))
	b.ln()
	b.print("%%EOF\n")
}
