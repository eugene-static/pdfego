package primitives

import "github.com/eugene-static/pdfego/unit"

const (
	LiteralStringOpen      = '('
	LiteralStringClose     = ')'
	HexadecimalStringOpen  = '<'
	HexadecimalStringClose = '>'
)

type String []unit.HEX

func (s String) Append(dst []byte) []byte {
	dst = append(dst, HexadecimalStringOpen)

	for _, hex := range s {
		dst = hex.Append(dst)
	}

	dst = append(dst, HexadecimalStringClose)

	return dst
}
