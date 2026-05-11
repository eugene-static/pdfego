package template

import (
	"bytes"
	"slices"

	"golang.org/x/image/font/sfnt"
)

type buffer struct {
	content *bytes.Buffer
}

func newBuffer() *buffer {
	return &buffer{
		content: new(bytes.Buffer),
	}
}

func (b *buffer) reset() {
	b.content.Reset()
}

func (b *buffer) writeFrom(buf *bytes.Buffer) {
	buf.WriteTo(b.content) //TODO: обработка ошибок
	b.ln()
}

func (b *buffer) write(data []byte) {
	b.content.Write(data)
	b.ln()
}

func (b *buffer) print(s ...string) {
	for i := range s {
		b.content.WriteString(s[i])
	}
}

func (b *buffer) printText(font *font, text string) {
	b.content.WriteByte('<')
	defer b.content.WriteByte('>')

	for _, r := range text {
		gl := font.glyph(r)

		writeUint16D4(b.content, uint16(gl.index))
	}
}

func (b *buffer) printFloat64(v float64) {
	writeFloat64(b.content, v)
}

func (b *buffer) printInt64(v int64) {
	writeInt64(b.content, v)
}

func (b *buffer) printUint16(v uint16) {
	writeUint16(b.content, v)
}

func (b *buffer) space() {
	b.content.WriteByte(' ')
}

func (b *buffer) ln() {
	b.content.WriteByte('\n')
}

// /REG 14 Tf
func (b *buffer) printFont(alias string, fontSize int) {
	b.print("/", alias, " ")
	b.printFloat64(float64(fontSize))
	b.print(" Tf\n")
}

// /W [1 [100] 3 [95 83 99]]
func (b *buffer) printGlyphWidthTable(glyphs []*glyph) {
	b.print("/W [")

	prev := sfnt.GlyphIndex(0)
	for _, gl := range glyphs {
		if gl.index == prev+1 {
			b.print(" ")
			b.printInt64(int64(gl.advance))

			continue
		}

		if prev > 0 {
			b.print("] ")
		}

		b.printUint16(uint16(gl.index))
		b.print(" [")
		b.printInt64(int64(gl.advance))

		prev = gl.index
	}

	b.print("]]\n")
}

func (b *buffer) printGlyphCharDictionary(glyphs []*glyph) {
	for chunk := range slices.Chunk(glyphs, 100) {
		b.printInt64(int64(len(chunk)))
		b.print(" beginbfchar\n")

		for _, gl := range chunk {
			b.print("<")
			writeUint16D4(b.content, uint16(gl.index))
			b.print("> <")
			writeUint16D4(b.content, gl.rune)
			b.print(">\n")
		}

		b.print("endbfchar\n")
	}
}

// 0 0
func (b *buffer) printXY(x, y float64) {
	b.printFloat64(x)
	b.space()
	b.printFloat64(y)
}

// 1 0 obj
func (b *buffer) startObj(objNum int64) {
	b.printInt64(objNum)
	b.content.WriteString(" 0 obj\n")
}

// endobj
func (b *buffer) endObj() {
	b.content.WriteString("endobj\n\n")
}

// <<
func (b *buffer) openObjectParameters() {
	b.content.WriteString("<<\n")
}

// >>
func (b *buffer) closeObjectParameters() {
	b.content.WriteString(">>\n")
}

// stream
func (b *buffer) startStream() {
	b.content.WriteString("stream\n")
}

// endstream
func (b *buffer) endStream() {
	b.content.WriteString("endstream\n")
}

// /Parent 1 0 R
func (b *buffer) printRef(field string, objNum int64) {
	b.content.WriteString(field)
	b.space()
	b.printInt64(objNum)
	b.content.WriteString(" 0 R\n")
}

// /Kids [2 0 R 3 0 R]
func (b *buffer) printRefArray(field string, objNums []int64) {
	b.print(field, " [")

	for i := range objNums {
		b.printInt64(objNums[i])
		b.print(" 0 R")
		if i < len(objNums)-1 {
			b.space()
		}
	}

	b.print("]\n")
}

func (b *buffer) printXref(ref int) {
	writeInt64D10(b.content, int64(ref))
	b.print(" 00000 n\n")
}

// /Flag 4
func (b *buffer) printFieldInt(field string, value int) {
	b.print(field, " ")
	b.printInt64(int64(value))
	b.ln()
}

// /Type /Page
func (b *buffer) printFieldString(field, value string) {
	b.print(field, " ", value, "\n")
}

func (b *buffer) printFieldStringWithBrackets(field, value string) {
	b.print(field, " (", value, ")\n")
}

// /FontBBox [-50 -200 800 800]
func (b *buffer) printFieldIntArray(field string, arr []int) {
	b.print(field, " [")

	for i := range arr {
		b.printInt64(int64(arr[i]))
		if i < len(arr)-1 {
			b.space()
		}
	}

	b.print("]\n")
}

// /MediaBox [0 0 595.28 841.89]
func (b *buffer) printFieldFloatArray(field string, arr []float64) {
	b.print(field, " [")

	for i := range arr {
		b.printFloat64(arr[i])
		if i < len(arr)-1 {
			b.space()
		}
	}

	b.print("]\n")
}

// 1 w x0 y0 m x1 y1 l S
func (b *buffer) printLine(bw, x0, y0, x1, y1 float64) {
	b.printFloat64(bw)
	b.print(" w ")
	b.printXY(x0, y0)
	b.print(" m ")
	b.printXY(x1, y1)
	b.print(" l S ")
}

// 1 w x0 y0 w h re S
func (b *buffer) printRect(bw, x, y, w, h float64) {
	b.printFloat64(bw)
	b.print(" w ")
	b.printXY(x, y)
	b.space()
	b.printXY(w, -h)
	b.print(" re S ")
}
