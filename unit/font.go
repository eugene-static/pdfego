package unit

import (
	"unicode/utf16"

	"golang.org/x/image/math/fixed"
)

const hexTable = "0123456789ABCDEF"

type Font int32

func (fu Font) EM(upm fixed.Int26_6) EM {
	return EM(fu) / EM(upm)
}

func (fu Font) Round() int {
	return int((fu + 0x20) >> 6)
}

type EM float64

func (em EM) PT(size PT) PT {
	return PT(em) * size
}

func (em EM) MM(size PT) MM {
	return (PT(em) * size).MM()
}

func (em EM) Font(upm fixed.Int26_6) Font {
	return Font(em * EM(upm.Round()))
}

type HEX uint16

func (hex HEX) Append(dst []byte) []byte {
	return append(dst,
		hexTable[(hex>>12)&0xF],
		hexTable[(hex>>8)&0xF],
		hexTable[(hex>>4)&0xF],
		hexTable[hex&0xF],
	)
}

type Rune int32

func (r Rune) Append(dst []byte) []byte {
	if r <= 0xFFFF {
		return HEX(r).Append(dst)
	}

	r1, r2 := utf16.EncodeRune(rune(r))

	dst = HEX(r1).Append(dst)
	dst = HEX(r2).Append(dst)

	return dst
}
