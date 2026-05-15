package core

import (
	"log/slog"
	"time"

	"github.com/eugene-static/pdf-craft/buffer"
	"github.com/eugene-static/pdf-craft/font"
)

const (
	pagesObjNum     = 1
	resourcesObjNum = 2
)

func (core *Core) FillBuffer() {
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
			alias:  f.Alias(),
			objNum: fontObjNum,
		})
	}

	b := core.mainBuffer

	core.setObject(resourcesObjNum)

	b.StartObj(resourcesObjNum)
	b.OpenObjectParameters()
	b.WriteFieldString("/Font", "")
	b.OpenObjectParameters()

	for _, resource := range fontResources {
		b.WriteRef("/"+resource.alias, resource.objNum)
	}

	b.CloseObjectParameters()
	b.CloseObjectParameters()
	b.EndObj()
}

func (core *Core) writeFont(f *font.Font) int64 {
	b := core.mainBuffer
	alias := "/" + f.Alias()

	cMapB := buffer.New()

	glyphs := f.Glyphs()

	// /CIDInit /ProcSet findresource begin 12 dict begin begincmap /CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def /CMapName /Adobe-Identity-UCS def /CMapType 2 def
	// 1 begincodespacerange <0000> <FFFF> endcodespacerange
	// 2 beginbfchar
	// <01CE> <0434>
	// <01D8> <043E>
	// endbfchar endcmap /CMapName currentdict /CMap defineresource pop end end
	cMapB.WriteFieldString("/CIDInit", "/ProcSet findresource begin")
	cMapB.WriteFieldString("12", "dict begin")
	cMapB.WriteStringLn("begincmap")
	cMapB.WriteFieldString("/CIDSystemInfo", "<< /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def")
	cMapB.WriteFieldString("/CMapName", "/Adobe-Identity-UCS def")
	cMapB.WriteFieldString("/CMapType", "2 def")
	cMapB.WriteStringLn("1 begincodespacerange")
	cMapB.WriteStringLn("<0000> <FFFF>")
	cMapB.WriteStringLn("endcodespacerange")
	cMapB.WriteGlyphCharDictionary(glyphs)
	cMapB.WriteStringLn("endcmap")
	cMapB.WriteFieldString("/CMapName", "currentdict")
	cMapB.WriteFieldString("/CMap", "defineresource pop")
	cMapB.WriteStringLn("end\nend")

	// ---------- (CMap) ----------
	// "8 0 obj << /Length %length >> stream ... endstream endobj\n"
	cMapOnjNum := core.newObject()

	b.StartObj(cMapOnjNum)
	b.OpenObjectParameters()
	b.WriteFieldInt("/Length", cMapB.Len())
	b.CloseObjectParameters()
	b.StartStream()
	b.WriteFrom(cMapB)
	b.EndStream()
	b.EndObj()

	// ---------- (Type0) ----------
	// "9 0 obj<< /Type /Font /Subtype /Type0 /BaseFont /%font_name /Encoding /Identity-H /DescendantFonts [10 0 R] >>endobj\n"
	fontNum := core.newObject()

	b.StartObj(fontNum)
	b.OpenObjectParameters()
	b.WriteFieldString("/Type", "/Font")
	b.WriteFieldString("/Subtype", "/Type0")
	b.WriteFieldString("/BaseFont", alias)
	b.WriteFieldString("/Encoding", "/Identity-H")
	b.WriteRefArray("/DescendantFonts", []int64{fontNum + 1})
	b.WriteRef("/ToUnicode", cMapOnjNum)
	b.CloseObjectParameters()
	b.EndObj()

	// ---------- (CIDFontType2) ----------
	// "10 0 obj<< /Type /Font /Subtype /CIDFontType2 /BaseFont /%font_name /CIDSystemInfo<< /Registry(Adobe) /Ordering(Identity) /Supplement 0 >> /FontDescriptor 11 0 R /DW 1000 >>endobj\n"
	objNum := core.newObject()

	b.StartObj(objNum)
	b.OpenObjectParameters()
	b.WriteFieldString("/Type", "/Font")
	b.WriteFieldString("/Subtype", "/CIDFontType2")
	b.WriteFieldString("/BaseFont", alias)
	b.WriteFieldString("/CIDSystemInfo", "<< /Registry(Adobe) /Ordering(Identity) /Supplement 0 >>")
	b.WriteRef("/FontDescriptor", objNum+1)
	b.WriteFieldInt("/DW", 600)
	b.WriteGlyphWidthTable(glyphs)
	b.CloseObjectParameters()
	b.EndObj()

	// ---------- (FontDescriptor) ----------
	// "11 0 obj<< /Type /FontDescriptor /FontName /%font_name /Flags 4 /FontBBox[-50 -200 800 800] /ItalicAngle 0 /Ascent 800 /Descent -200 /CapHeight 700 /StemV 80 /FontFile2 12 0 R >>endobj\n"
	objNum = core.newObject()
	metrics := f.Metrics()

	b.StartObj(objNum)
	b.OpenObjectParameters()
	b.WriteFieldString("/Type", "/FontDescriptor")
	b.WriteFieldString("/FontName", alias)
	b.WriteFieldInt("/Flags", 4)
	b.WriteFieldIntArray("/FontBBox", metrics.FontBBox)
	b.WriteFieldInt("/ItalicAngle", 0) //TODO: Italic Font
	b.WriteFieldInt("/Ascent", metrics.Ascent)
	b.WriteFieldInt("/Descent", metrics.Descent)
	b.WriteFieldInt("/CapHeight", metrics.CapHeight)
	b.WriteFieldInt("/StemV", metrics.StemV)
	b.WriteRef("/FontFile2", objNum+1)
	b.CloseObjectParameters()
	b.EndObj()

	// ---------- (FontFile2) ----------
	// "12 0 obj<< /Length %font_bytes_length /Length1 %font_bytes_length >>stream\nfont_bytes\nendstream\nendobj\n"
	objNum = core.newObject()

	data, l, err := f.Data()
	if err != nil {
		core.log.Debug("error reading data from font", slog.String("err", err.Error()))
	}

	b.StartObj(objNum)
	b.OpenObjectParameters()
	b.WriteFieldString("/Filter", "/FlateDecode")
	b.WriteFieldInt("/Length", len(data))
	b.WriteFieldInt("/Length1", l)
	b.CloseObjectParameters()
	b.StartStream()
	b.Write(data)
	b.EndStream()
	b.EndObj()

	return fontNum
}

func (core *Core) writePages() {
	pageObjs := make([]int64, 0, len(core.pageBuffers))

	for _, buf := range core.pageBuffers {
		pageObj := core.writePage(buf)

		pageObjs = append(pageObjs, pageObj)
	}

	b := core.mainBuffer

	core.setObject(pagesObjNum)

	b.StartObj(pagesObjNum)
	b.OpenObjectParameters()
	b.WriteFieldString("/Type", "/Pages")
	b.WriteRefArray("/Kids", pageObjs)
	b.WriteFieldInt("/Count", len(pageObjs))
	b.WriteFieldFloatArray("/MediaBox", []float64{0, 0, core.page.width, core.page.height})
	b.CloseObjectParameters()
	b.EndObj()
}

func (core *Core) writePage(buf *buffer.Buffer) int64 {
	b := core.mainBuffer
	pageObjNum := core.newObject()

	b.StartObj(pageObjNum)
	b.OpenObjectParameters()
	b.WriteFieldString("/Type", "/Page")
	b.WriteRef("/Parent", 1)
	b.WriteRef("/Resources", 2)
	b.WriteRef("/Contents", pageObjNum+1)
	b.CloseObjectParameters()
	b.EndObj()

	objNum := core.newObject()

	b.StartObj(objNum)
	b.OpenObjectParameters()
	b.WriteFieldInt("/Length", buf.Len())
	b.CloseObjectParameters()
	b.StartStream()
	b.WriteFrom(buf)
	b.EndStream()
	b.EndObj()

	return pageObjNum
}

func (core *Core) writeFileHeader() {
	core.mainBuffer.WriteStringLn("%PDF-1.4")
	core.mainBuffer.WriteStringLn("%\x80\x80\x80\x80")
}

func (core *Core) writeInfo() int64 {
	creationDate := time.Now().Format("D:20060102150405-07'00'")

	b := core.mainBuffer
	objNum := core.newObject()

	b.StartObj(objNum)
	b.OpenObjectParameters()
	b.WriteFieldStringWithBrackets("/Producer", "OVP")
	b.WriteFieldStringWithBrackets("/CreationDate", creationDate)
	b.CloseObjectParameters()
	b.EndObj()

	return objNum
}

func (core *Core) writeCatalog() int64 {
	b := core.mainBuffer
	objNum := core.newObject()

	b.StartObj(objNum)
	b.OpenObjectParameters()
	b.WriteFieldString("/Type", "/Catalog")
	b.WriteRef("/Pages", pagesObjNum)
	b.CloseObjectParameters()
	b.EndObj()

	return objNum
}

func (core *Core) writeXref() int {
	xrefOffset := core.mainBuffer.Len()

	b := core.mainBuffer

	b.WriteStringLn("xref")
	b.WriteFieldInt("0", len(core.offsets))
	b.WriteStringLn("0000000000 65535 f")
	for i := range core.offsets {
		if i == 0 {
			continue
		}

		b.WriteXref(core.offsets[i])
	}

	return xrefOffset
}

func (core *Core) writeTrailer(root, info int64) {
	b := core.mainBuffer

	b.WriteStringLn("trailer")
	b.OpenObjectParameters()
	b.WriteFieldInt("/Size", len(core.offsets))
	b.WriteRef("/Root", root)
	b.WriteRef("/Info", info)
	b.CloseObjectParameters()
}

func (core *Core) writeEOF(xrefOffset int) {
	b := core.mainBuffer

	b.WriteFieldInt("startxref\n", xrefOffset)
	b.WriteStringLn("%%EOF")
}
