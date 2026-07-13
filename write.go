package pdfego

import (
	"strconv"
	"time"

	"github.com/eugene-static/pdfego/internal/buffer"
	"github.com/eugene-static/pdfego/internal/font"
	"github.com/eugene-static/pdfego/internal/image"
)

const (
	objNumPages     = 1
	objNumResources = 2
)

func (core *Core) startDocument() {
	core.writeFileHeader()
}

func (core *Core) finishDocument() {
	core.writePages()
	core.writeResources()

	infoObj := core.writeInfo()
	rootObj := core.writeCatalog()
	xrefOffset := core.writeXref()

	core.writeTrailer(rootObj, infoObj)
	core.writeEOF(xrefOffset)
}

func (core *Core) writeResources() {
	if core.err() != nil {
		return
	}

	type resource struct {
		alias  string
		objNum int
	}

	fontResources := make([]resource, 0, len(core.fonts))
	imageResources := make([]resource, 0, len(core.images))

	for alias, f := range core.fonts {
		fontObjNum := core.writeFont(f, alias)

		fontResources = append(fontResources, resource{
			alias:  alias,
			objNum: fontObjNum,
		})
	}

	for alias, img := range core.images {
		imageObjNum := core.writeImage(img)

		imageResources = append(imageResources, resource{
			alias:  alias,
			objNum: imageObjNum,
		})
	}

	b := core.mainBuffer

	core.setObject(objNumResources)

	b.StartObj(objNumResources)
	b.OpenObjectParameters()

	b.WriteFieldString("/Font", "")
	b.OpenObjectParameters()

	for _, res := range fontResources {
		b.WriteRef("/"+res.alias, res.objNum)
	}

	b.CloseObjectParameters()
	b.WriteFieldString("/XObject", "")
	b.OpenObjectParameters()

	for _, res := range imageResources {
		b.WriteRef("/"+res.alias, res.objNum)
	}

	b.CloseObjectParameters()

	b.CloseObjectParameters()
	b.EndObj()
}

func (core *Core) writeFont(f *font.Font, alias string) int {
	if core.err() != nil {
		return 0
	}

	b := core.mainBuffer
	alias = "/" + alias

	cMapB := buffer.New(1 << 10)

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
	cMapObjNum := core.newObject()

	b.StartObj(cMapObjNum)
	b.OpenObjectParameters()
	b.WriteFieldInt("/Length", cMapB.Len())
	b.CloseObjectParameters()
	b.StartStream()

	_, err := b.ReadFrom(cMapB)
	if err != nil {
		core.setError(err)

		return 0
	}

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
	b.WriteRefArray("/DescendantFonts", []int{fontNum + 1})
	b.WriteRef("/ToUnicode", cMapObjNum)
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
	b.WriteFieldInt("/CapHeight", metrics.CapHeight.Round())
	b.WriteFieldInt("/StemV", metrics.StemV)
	b.WriteRef("/FontFile2", objNum+1)
	b.CloseObjectParameters()
	b.EndObj()

	// ---------- (FontFile2) ----------
	// "12 0 obj<< /Length %font_bytes_length /Length1 %font_bytes_length >>stream\nfont_bytes\nendstream\nendobj\n"
	objNum = core.newObject()

	fontRawBytes := f.Bytes()

	fontBytes, ok := f.CompressedBytes()
	if !ok {
		subset, err := f.Subset()
		if err != nil {
			core.setError(err)

			return 0
		}

		compressedBytes, err := core.comp.compress(subset)
		if err != nil {
			core.setError(err)

			return 0
		}

		f.SaveCompressedBytes(compressedBytes)

		fontBytes = compressedBytes
	}

	b.StartObj(objNum)
	b.OpenObjectParameters()
	b.WriteFieldString("/Filter", "/FlateDecode")
	b.WriteFieldInt("/Length", len(fontBytes))
	b.WriteFieldInt("/Length1", len(fontRawBytes))
	b.CloseObjectParameters()
	b.StartStream()

	_, err = b.Write(fontBytes)
	if err != nil {
		return 0
	}

	b.EndStream()
	b.EndObj()

	return fontNum
}

func (core *Core) writeImage(img *image.Image) int {
	if core.err() != nil {
		return 0
	}

	b := core.mainBuffer

	var alphaObjNum int

	alphaBytes, compressed := img.Alpha()
	if alphaBytes != nil {
		alphaObjNum = core.newObject()

		if !compressed {
			compBytes, err := core.comp.compress(alphaBytes)
			if err != nil {
				core.setError(err)

				return 0
			}

			img.SaveCompressedAlpha(compBytes)

			alphaBytes = compBytes
		}

		b.StartObj(alphaObjNum)
		b.OpenObjectParameters()
		b.WriteFieldString("/Type", "/XObject")
		b.WriteFieldString("/Subtype", "/Image")
		b.WriteFieldInt("/Width", img.Width())
		b.WriteFieldInt("/Height", img.Height())
		b.WriteFieldString("/ColorSpace", "/DeviceGray")
		b.WriteFieldInt("/BitsPerComponent", 8)
		b.WriteFieldString("/Filter", "/FlateDecode")
		b.WriteFieldInt("/Length", len(alphaBytes))
		b.CloseObjectParameters()
		b.StartStream()

		_, err := b.Write(alphaBytes)
		if err != nil {
			core.setError(err)

			return 0
		}

		b.EndStream()
		b.EndObj()
	}

	objNum := core.newObject()

	imageBytes, compressed := img.RGB()
	if !compressed {
		compressedBytes, err := core.comp.compress(imageBytes)
		if err != nil {
			core.setError(err)

			return 0
		}

		img.SaveCompressedRGB(compressedBytes)

		imageBytes = compressedBytes
	}

	b.StartObj(objNum)
	b.OpenObjectParameters()
	b.WriteFieldString("/Type", "/XObject")
	b.WriteFieldString("/Subtype", "/Image")
	b.WriteFieldInt("/Width", img.Width())
	b.WriteFieldInt("/Height", img.Height())
	b.WriteFieldString("/ColorSpace", "/DeviceRGB")
	b.WriteFieldInt("/BitsPerComponent", 8)
	b.WriteFieldString("/Filter", "/FlateDecode")
	b.WriteFieldInt("/Length", len(imageBytes))

	if alphaObjNum > 0 {
		b.WriteRef("/SMask", alphaObjNum)
	}

	b.CloseObjectParameters()
	b.StartStream()

	_, err := b.Write(imageBytes)
	if err != nil {
		core.setError(err)

		return 0
	}

	b.EndStream()
	b.EndObj()

	return objNum
}

func (core *Core) writeWatermark() {
	if core.err() != nil {
		return
	}

	b := core.mainBuffer
	objNum := core.newObject()

	b.StartObj(objNum)
	b.OpenObjectParameters()

	watermarkBytes := core.page.watermark.buffer.Bytes()
	length := core.page.watermark.buffer.Len()

	if core.compress {
		compressed, err := core.comp.compress(watermarkBytes)
		if err != nil {
			core.setError(err)

			return
		}

		watermarkBytes = compressed
		length = len(compressed)
		b.WriteFieldString("/Filter", "/FlateDecode")
	}

	b.WriteFieldInt("/Length", length)
	b.CloseObjectParameters()
	b.StartStream()

	_, err := b.Write(watermarkBytes)
	if err != nil {
		core.setError(err)

		return
	}

	b.EndStream()
	b.EndObj()

	core.page.watermark.objNum = objNum
}

func (core *Core) writePages() {
	if core.err() != nil {
		return
	}

	b := core.mainBuffer

	core.setObject(objNumPages)

	b.StartObj(objNumPages)
	b.OpenObjectParameters()
	b.WriteFieldString("/Type", "/Pages")
	b.WriteRefArray("/Kids", core.page.objects)
	b.WriteFieldInt("/Count", len(core.page.objects))
	b.WriteFieldFloatArray("/MediaBox", []float64{0, 0, core.page.width.PT().Float64(), core.page.height.PT().Float64()})
	b.CloseObjectParameters()
	b.EndObj()
}

func (core *Core) writeUpdatedPages() {
	if core.err() != nil {
		return
	}

	b := core.mainBuffer

	objNum := core.newObject()

	b.StartObj(objNum)

}

func (core *Core) writePage() {
	if core.err() != nil {
		return
	}

	contentsObjNum := core.writeContents()
	pageObjNum := core.newObject()

	b := core.mainBuffer

	contents := []int{contentsObjNum}
	if core.page.watermark.objNum > 0 {
		contents = append(contents, core.page.watermark.objNum)
	}

	b.StartObj(pageObjNum)
	b.OpenObjectParameters()
	b.WriteFieldString("/Type", "/Page")
	b.WriteRef("/Parent", 1)
	b.WriteRef("/Resources", 2)
	b.WriteRefArray("/Contents", contents)
	b.CloseObjectParameters()
	b.EndObj()

	core.page.objects = append(core.page.objects, pageObjNum)
}

//func (core *Core) writeUpdatedPage(page *object, contents []int) {
//	if core.err() != nil {
//		return
//	}
//
//	type param struct {
//		name  string
//		value parameter
//	}
//
//	params := make([]param, 0, len(page.parameters))
//
//	for name, value := range page.parameters {
//		params = append(params, param{name, value})
//	}
//
//	sort.Slice(params, func(i, j int) bool {
//		iStart, _ := params[i].value.bounds()
//		jStart, _ := params[j].value.bounds()
//
//		return iStart < jStart
//	})
//
//	b := core.mainBuffer
//	pageObjNum := core.newObject()
//
//	b.StartObj(pageObjNum)
//	b.OpenObjectParameters()
//
//	pos := 0
//
//	for _, p := range params {
//		start, end := p.value.bounds()
//
//		if start < pos {
//			continue
//		}
//
//		if start > len(page.body) {
//			break
//		}
//
//		if end > len(page.body) {
//			end = len(page.body)
//		}
//
//		b.Write(page.body[pos:start])
//
//		if p.value.title() == "Contents" {
//
//		}
//	}
//}

func (core *Core) writeContents() int {
	if core.err() != nil {
		return 0
	}

	b := core.mainBuffer

	objNum := core.newObject()

	b.StartObj(objNum)
	b.OpenObjectParameters()

	pageBytes := core.page.buffer.Bytes()
	length := core.page.buffer.Len()

	if core.compress {
		compressed, err := core.comp.compress(pageBytes)
		if err != nil {
			core.setError(err)

			return 0
		}

		pageBytes = compressed
		length = len(compressed)
		b.WriteFieldString("/Filter", "/FlateDecode")
	}

	b.WriteFieldInt("/Length", length)
	b.CloseObjectParameters()
	b.StartStream()

	_, err := b.Write(pageBytes)
	if err != nil {
		core.setError(err)

		return 0
	}

	b.EndStream()
	b.EndObj()

	return objNum
}

func (core *Core) writeFileHeader() {
	if core.err() != nil {
		return
	}

	core.mainBuffer.WriteStringLn("%PDF-1.6")
	core.mainBuffer.WriteStringLn("%\x80\x80\x80\x80")
}

func (core *Core) writeInfo() int {
	if core.err() != nil {
		return 0
	}

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

func (core *Core) writeCatalog() int {
	if core.err() != nil {
		return 0
	}

	b := core.mainBuffer
	objNum := core.newObject()

	b.StartObj(objNum)
	b.OpenObjectParameters()
	b.WriteFieldString("/Type", "/Catalog")
	b.WriteRef("/Pages", objNumPages)
	b.CloseObjectParameters()
	b.EndObj()

	return objNum
}

func (core *Core) writeXref() int {
	if core.err() != nil {
		return 0
	}

	xrefOffset := core.mainBuffer.Len()

	b := core.mainBuffer

	b.WriteStringLn("xref")
	b.WriteFieldInt("0", len(core.offsets))
	b.WriteStringLn("0000000000 65535 f\r")
	for i := range core.offsets {
		if i == 0 {
			continue
		}

		b.WriteXref(core.offsets[i])
	}

	return xrefOffset
}

func (core *Core) writeTrailer(root, info int) {
	if core.err() != nil {
		return
	}

	b := core.mainBuffer

	b.WriteStringLn("trailer")
	b.OpenObjectParameters()
	b.WriteFieldInt("/Size", len(core.offsets))
	b.WriteRef("/Root", root)
	b.WriteRef("/Info", info)
	b.CloseObjectParameters()
}

func (core *Core) writeEOF(xrefOffset int) {
	if core.err() != nil {
		return
	}

	b := core.mainBuffer

	b.WriteStringLn("startxref")
	b.WriteStringLn(strconv.Itoa(xrefOffset))
	b.WriteStringLn("%%EOF")
}
