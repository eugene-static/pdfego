package template

import (
	"bytes"
	"strconv"
)

const (
	hex  = "0123456789ABCDEF"
	dpi  = 72.0
	inch = 25.4
)

func appendFloat(b *bytes.Buffer, val float64) {
	buf := b.AvailableBuffer()
	buf = strconv.AppendFloat(buf, val, 'f', 2, 64)
	b.Write(buf)
}

func appendInt(b *bytes.Buffer, val int64) {
	buf := b.AvailableBuffer()
	buf = strconv.AppendInt(buf, val, 10)
	b.Write(buf)
}

func appendInt10(b *bytes.Buffer, val int64) {
	// Преобразуем в десятичное представление
	var tmp [12]byte // достаточно для int под 32/64
	buf := strconv.AppendInt(tmp[:0], val, 10)
	// Дописываем нули впереди до длины 10
	zeros := 10 - len(buf)
	for zeros > 0 {
		b.WriteByte('0')
		zeros--
	}

	b.Write(buf)
}

func appendHex4(b *bytes.Buffer, val uint16) {
	buf := b.AvailableBuffer()
	buf = append(buf,
		//'<',
		hex[(val>>12)&0xF],
		hex[(val>>8)&0xF],
		hex[(val>>4)&0xF],
		hex[val&0xF],
		//'>',
	)

	b.Write(buf)
}

// Возвращает пункты (кегль) в миллиметрах.
func pt(v float64) float64 {
	return v / (dpi / inch)
}

// Возвращает миллиметры в пунктах
func mm(v float64) float64 {
	return v * (dpi / inch)
}
