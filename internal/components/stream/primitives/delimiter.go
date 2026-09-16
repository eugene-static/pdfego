package primitives

const (
	ArrayOpen  = '['
	ArrayClose = ']'

	LiteralStringOpen      = '('
	LiteralStringClose     = ')'
	HexadecimalStringOpen  = '<'
	HexadecimalStringClose = '>'

	DictionaryOpen  Delimiter = "<<"
	DictionaryClose Delimiter = ">>"
)

type Delimiter string

func (d Delimiter) Append(dst []byte) []byte {
	return append(dst, d...)
}
