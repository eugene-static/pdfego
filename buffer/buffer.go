package buffer

import (
	"bytes"
	"slices"
	"strconv"

	"github.com/eugene-static/pdf-craft/font"
	"github.com/eugene-static/pdf-craft/meter"
)

type Buffer struct {
	content *bytes.Buffer
}

func New(n int) *Buffer {
	return &Buffer{
		content: bytes.NewBuffer(make([]byte, 0, n)),
	}
}

func (b *Buffer) Len() int {
	return b.content.Len()
}

func (b *Buffer) Cap() int {
	return b.content.Cap()
}

func (b *Buffer) Grow(n int) {
	b.content.Grow(n)
}

func (b *Buffer) Bytes() []byte {
	return b.content.Bytes()
}

func (b *Buffer) Reset() {
	b.content.Reset()
}

func (b *Buffer) ReadFrom(buf *Buffer) {
	bufBytes := buf.Bytes()

	//fmt.Printf("len(bufBytes)=%d\n", len(bufBytes))
	b.content.Grow(len(bufBytes) + 1)
	b.content.Write(bufBytes) //TODO: обработка ошибок
	buf.Reset()
	b.ln()
}

func (b *Buffer) Write(data []byte) (int, error) {
	b.content.Grow(len(data) + 1)
	b.content.Write(data)
	b.ln()

	return len(data) + 1, nil
}

func (b *Buffer) WriteStringLn(val string) {
	b.content.Grow(len(val))
	b.content.WriteString(val)
	b.ln()
}

// /REG 14 Tf
func (b *Buffer) WriteFont(alias string, fontSize meter.PT) {
	b.writeString("/", alias, " ")
	b.writeFloat64(fontSize.Float64())
	b.writeString(" Tf\n")
}

// "1 0 0 1 x y Tm" задает абсолютную позицию текста на странице.
func (b *Buffer) WriteText(font *font.Font, x, y meter.MM, text string) {
	b.content.WriteString("1 0 0 1 ")
	b.writeXY(x, y)
	b.content.WriteString(" Tm <")

	for _, r := range text {
		gid := font.GID(r)
		b.writeUint16D4(gid)
	}

	b.content.WriteString("> Tj\n")
}

// /W [1 [100] 3 [95 83 99]]
func (b *Buffer) WriteGlyphWidthTable(glyphs []font.Glyph) {
	b.writeString("/W [")

	prev := uint16(0)
	for _, gl := range glyphs {
		index := gl.Index()
		advance := gl.Advance()

		if index == prev+1 {
			b.writeString(" ")
			b.writeInt64(advance)

			continue
		}

		if prev > 0 {
			b.writeString("] ")
		}

		b.writeUint16(index)
		b.writeString(" [")
		b.writeInt64(advance)

		prev = index
	}

	b.writeString("]]\n")
}

func (b *Buffer) WriteGlyphCharDictionary(glyphs []font.Glyph) {
	for chunk := range slices.Chunk(glyphs, 100) {
		b.writeInt64(int64(len(chunk)))
		b.writeString(" beginbfchar\n")

		for _, gl := range chunk {
			b.writeString("<")
			b.writeUint16D4(gl.Index())
			b.writeString("> <")
			b.writeUint16D4(gl.Rune())
			b.writeString(">\n")
		}

		b.writeString("endbfchar\n")
	}
}

// StartObj writes "1 0 obj" to buffer.
func (b *Buffer) StartObj(objNum int64) {
	b.writeInt64(objNum)
	b.content.WriteString(" 0 obj\n")
}

// endobj
func (b *Buffer) EndObj() {
	b.content.WriteString("endobj\n\n")
}

// <<
func (b *Buffer) OpenObjectParameters() {
	b.content.WriteString("<<\n")
}

// >>
func (b *Buffer) CloseObjectParameters() {
	b.content.WriteString(">>\n")
}

// stream
func (b *Buffer) StartStream() {
	b.content.WriteString("stream\n")
}

// endstream
func (b *Buffer) EndStream() {
	b.content.WriteString("endstream\n")
}

// /Parent 1 0 R
func (b *Buffer) WriteRef(field string, objNum int64) {
	b.writeString(field)
	b.space()
	b.writeInt64(objNum)
	b.writeString(" 0 R\n")
}

// /Kids [2 0 R 3 0 R]
func (b *Buffer) WriteRefArray(field string, objNums []int64) {
	b.writeString(field, " [")

	for i := range objNums {
		b.writeInt64(objNums[i])
		b.writeString(" 0 R")
		if i < len(objNums)-1 {
			b.space()
		}
	}

	b.writeString("]\n")
}

func (b *Buffer) WriteXref(ref int) {
	b.writeInt64D10(int64(ref))
	b.writeString(" 00000 n\r\n")
}

// /Type /Page
func (b *Buffer) WriteFieldString(field, value string) {
	b.writeString(field, " ", value, "\n")
}

func (b *Buffer) WriteFieldStringWithBrackets(field, value string) {
	b.writeString(field, " (", value, ")\n")
}

// /Flag 4
func (b *Buffer) WriteFieldInt(field string, value int) {
	b.writeString(field, " ")
	b.writeInt64(int64(value))
	b.ln()
}

// /FontBBox [-50 -200 800 800]
func (b *Buffer) WriteFieldIntArray(field string, arr []int) {
	b.writeString(field, " [")

	for i := range arr {
		b.writeInt64(int64(arr[i]))
		if i < len(arr)-1 {
			b.space()
		}
	}

	b.writeString("]\n")
}

// /MediaBox [0 0 595.28 841.89]
func (b *Buffer) WriteFieldFloatArray(field string, arr []float64) {
	b.writeString(field, " [")

	for i := range arr {
		b.writeFloat64(arr[i])
		if i < len(arr)-1 {
			b.space()
		}
	}

	b.writeString("]\n")
}

// 1 w x0 y0 m x1 y1 l S
func (b *Buffer) WriteLine(bw meter.PT, x0, y0, x1, y1 meter.MM) {
	b.writeFloat64(bw.Float64())
	b.writeString(" w ")
	b.writeXY(x0, y0)
	b.writeString(" m ")
	b.writeXY(x1, y1)
	b.writeString(" l S ")
}

// 1 w x0 y0 w h re S
func (b *Buffer) WriteRect(bw meter.PT, x, y, w, h meter.MM) {
	b.writeFloat64(bw.Float64())
	b.writeString(" w ")
	b.writeXY(x, y)
	b.space()
	b.writeWH(w, h)
	b.writeString(" re S ")
}

func (b *Buffer) write(data []byte) {
	b.content.Grow(len(data))
	b.content.Write(data)

	return
}

func (b *Buffer) writeString(s ...string) {
	for i := range s {
		b.content.Grow(len(s[i]))
		b.content.WriteString(s[i])
	}
}

func (b *Buffer) writeFloat64(val float64) {
	buf := b.content.AvailableBuffer()
	buf = strconv.AppendFloat(buf, val, 'f', 2, 64)

	b.write(buf)
}

func (b *Buffer) writeInt64(val int64) {
	buf := b.content.AvailableBuffer()
	buf = strconv.AppendInt(buf, val, 10)

	b.write(buf)
}

func (b *Buffer) writeUint16(val uint16) {
	buf := b.content.AvailableBuffer()
	buf = strconv.AppendUint(buf, uint64(val), 10)

	b.write(buf)
}

func (b *Buffer) writeXY(x, y meter.MM) {
	b.writeFloat64(x.PT().Float64())
	b.space()
	b.writeFloat64(y.Abs().PT().Float64())
}

func (b *Buffer) writeWH(w, h meter.MM) {
	b.writeFloat64(w.PT().Float64())
	b.space()
	b.writeFloat64(h.Neg().PT().Float64())
}

const (
	hex = "0123456789ABCDEF"
)

func (b *Buffer) writeInt64D10(val int64) {
	const size = 10

	tmp := make([]byte, 0, size)

	buf := b.content.AvailableBuffer()
	buf = strconv.AppendInt(tmp[:0], val, 10)

	zeros := size - len(buf)

	for zeros > 0 {
		b.content.WriteByte('0')
		zeros--
	}

	b.write(buf)
}

func (b *Buffer) writeUint16D4(val uint16) {
	buf := b.content.AvailableBuffer()
	buf = append(buf,
		hex[(val>>12)&0xF],
		hex[(val>>8)&0xF],
		hex[(val>>4)&0xF],
		hex[val&0xF],
	)

	b.write(buf)
}

func (b *Buffer) space() {
	b.content.WriteByte(' ')
}

func (b *Buffer) ln() {
	b.content.WriteByte('\n')
}
