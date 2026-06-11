package buffer

import (
	"bytes"
	"slices"
	"strconv"

	"github.com/eugene-static/pdfego/internal/font"
	"github.com/eugene-static/pdfego/pkg/unit"
)

const DefaultSize = 1 << 16

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

func (b *Buffer) ReadFrom(buf *Buffer) (int, error) {
	bufBytes := buf.Bytes()
	lenBytes := len(bufBytes) + 1

	b.content.Grow(lenBytes)

	_, err := b.content.Write(bufBytes)
	if err != nil {
		return 0, err
	}

	buf.Reset()
	b.ln()

	return lenBytes, nil
}

func (b *Buffer) Write(data []byte) (int, error) {
	b.content.Grow(len(data))
	b.content.Write(data)

	return len(data), nil
}

func (b *Buffer) WriteStringLn(val string) {
	b.content.Grow(len(val))
	b.content.WriteString(val)
	b.ln()
}

// /$Alias $Size Tf
func (b *Buffer) WriteFont(alias string, fontSize unit.PT) {
	b.writeString("/", alias, " ")
	b.writeFloat64(fontSize.Float64())
	b.writeString(" Tf\n")
}

// 1 0 0 1 $X $Y Tm <$HEX1$HEX2...$HEXN> Tj
func (b *Buffer) WriteText(font *font.Font, x, y unit.MM, text string) {
	b.content.WriteString("1 0 0 1 ")
	b.writeXY(x, y)
	b.content.WriteString(" Tm <")

	for _, r := range text {
		gid := font.GID(r)
		b.writeUint16D4(gid)
	}

	b.content.WriteString("> Tj\n")
}

// $R $G $B rg
func (b *Buffer) WriteTextColor(red, green, blue float64) {
	b.writeFloat64(red)
	b.space()
	b.writeFloat64(green)
	b.space()
	b.writeFloat64(blue)
	b.WriteStringLn(" rg")
}

// $R $G $B RG
func (b *Buffer) WriteBorderColor(red, green, blue float64) {
	b.writeFloat64(red)
	b.space()
	b.writeFloat64(green)
	b.space()
	b.writeFloat64(blue)
	b.WriteStringLn(" RG")
}

// q $W 0 0 $H $X $Y cm /$ImageAlias Do Q
func (b *Buffer) WriteImage(x, y, w, h unit.MM, alias string) {
	b.content.WriteString("q ")
	b.writeFloat64(w.PT().Float64())
	b.writeString(" 0 0 ")
	b.writeFloat64(h.PT().Float64())
	b.space()
	b.writeXY(x, y)
	b.writeString(" cm /", alias, " Do Q\n")
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
			b.writeInt(advance)

			continue
		}

		if prev > 0 {
			b.writeString("] ")
		}

		b.writeUint16(index)
		b.writeString(" [")
		b.writeInt(advance)

		prev = index
	}

	b.writeString("]]\n")
}

func (b *Buffer) WriteGlyphCharDictionary(glyphs []font.Glyph) {
	for chunk := range slices.Chunk(glyphs, 100) {
		b.writeInt(len(chunk))
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

// $N 0 obj
func (b *Buffer) StartObj(objNum int) {
	b.writeInt(objNum)
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
	b.content.WriteString("\nendstream\n")
}

// /$Parent $N 0 R
func (b *Buffer) WriteRef(field string, objNum int) {
	b.writeString(field)
	b.space()
	b.writeInt(objNum)
	b.writeString(" 0 R\n")
}

// /Kids [$N1 0 R $N2 0 R]
func (b *Buffer) WriteRefArray(field string, objNums []int) {
	b.writeString(field, " [")

	for i := range objNums {
		b.writeInt(objNums[i])
		b.writeString(" 0 R")
		if i < len(objNums)-1 {
			b.space()
		}
	}

	b.writeString("]\n")
}

// 6500 00000 n
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
	b.writeInt(value)
	b.ln()
}

// /FontBBox [-50 -200 800 800]
func (b *Buffer) WriteFieldIntArray(field string, arr []int) {
	b.writeString(field, " [")

	for i := range arr {
		b.writeInt(arr[i])
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

// $Size w $X0 $Y0 m $X1 $Y1 l S
func (b *Buffer) WriteLine(bw unit.PT, x0, y0, x1, y1 unit.MM) {
	b.writeFloat64(bw.Float64())
	b.writeString(" w ")
	b.writeXY(x0, y0)
	b.writeString(" m ")
	b.writeXY(x1, y1)
	b.writeString(" l S\n")
}

// $Size w $X0 $Y0 $W $H re S
func (b *Buffer) WriteRect(bw unit.PT, x, y, w, h unit.MM) {
	b.writeFloat64(bw.Float64())
	b.writeString(" w ")
	b.writeXY(x, y)
	b.space()
	b.writeWH(w, h)
	b.writeString(" re S\n")
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

func (b *Buffer) writeInt(val int) {
	buf := b.content.AvailableBuffer()
	buf = strconv.AppendInt(buf, int64(val), 10)

	b.write(buf)
}

func (b *Buffer) writeUint16(val uint16) {
	buf := b.content.AvailableBuffer()
	buf = strconv.AppendUint(buf, uint64(val), 10)

	b.write(buf)
}

func (b *Buffer) writeXY(x, y unit.MM) {
	b.writeFloat64(x.PT().Float64())
	b.space()
	b.writeFloat64(y.Abs().PT().Float64())
}

func (b *Buffer) writeWH(w, h unit.MM) {
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
