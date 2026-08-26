package primitives

const (
	LiteralStringOpen      = '('
	LiteralStringClose     = ')'
	HexadecimalStringOpen  = '<'
	HexadecimalStringClose = '>'

	hexTable = "0123456789ABCDEF"
)

type HEX uint16

func (hex HEX) Append(dst []byte) []byte {
	return append(dst,
		hexTable[(hex>>12)&0xF],
		hexTable[(hex>>8)&0xF],
		hexTable[(hex>>4)&0xF],
		hexTable[hex&0xF],
	)
}

type String []HEX

func (s String) Append(dst []byte) []byte {
	dst = append(dst, HexadecimalStringOpen)

	for _, hex := range s {
		dst = hex.Append(dst)
	}

	dst = append(dst, HexadecimalStringClose)

	return dst
}
