package template

import (
	"bytes"
	"strconv"
	"strings"
)

const (
	hex  = "0123456789ABCDEF"
	dpi  = 72.0
	inch = 25.4
)

func writeFloat64(b *bytes.Buffer, val float64) {
	buf := b.AvailableBuffer()
	buf = strconv.AppendFloat(buf, val, 'f', 2, 64)

	b.Write(buf)
}

func writeInt64(b *bytes.Buffer, val int64) {
	buf := b.AvailableBuffer()
	buf = strconv.AppendInt(buf, val, 10)

	b.Write(buf)
}

func writeInt64D10(b *bytes.Buffer, val int64) {
	const size = 10
	tmp := make([]byte, 0, size)

	buf := b.AvailableBuffer()
	buf = strconv.AppendInt(tmp[:0], val, 10)

	zeros := size - len(buf)

	for zeros > 0 {
		b.WriteByte('0')
		zeros--
	}

	b.Write(buf)
}

func writeUint16(b *bytes.Buffer, val uint16) {
	buf := b.AvailableBuffer()
	buf = strconv.AppendUint(buf, uint64(val), 10)

	b.Write(buf)
}

func writeUint16D4(b *bytes.Buffer, val uint16) {
	buf := b.AvailableBuffer()
	buf = append(buf,
		hex[(val>>12)&0xF],
		hex[(val>>8)&0xF],
		hex[(val>>4)&0xF],
		hex[val&0xF],
	)

	b.Write(buf)
}

func splitText(f *font, text string, size int, width float64) []string {
	lines := make([]string, 0)

	for seg := range strings.Lines(text) {
		lines = append(lines, splitSegment(f, seg, size, width)...)
	}

	return lines
}

func splitSegment(f *font, text string, size int, width float64) []string {
	words := strings.Fields(text)

	lines := make([]string, 0, len(words))

	line := words[0]

	for _, word := range words[1:] {
		candidate := line + " " + word

		//TODO: хранить text_width, чтобы не считать заново при позиционировании

		candidateWidth := f.measureText(size, candidate)
		if candidateWidth > width {
			lines = append(lines, line)

			line = word

			continue
		}

		line = candidate
	}

	lines = append(lines, line)

	return lines
}

// Возвращает миллиметры в пунктах
func pt(v float64) float64 {
	return v * (dpi / inch)
}

// Возвращает пункты (кегль) в миллиметрах.
func mm(v float64) float64 {
	return v / (dpi / inch)
}
