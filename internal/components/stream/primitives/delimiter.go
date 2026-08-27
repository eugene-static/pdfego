package primitives

const (
	ArrayOpen                 = '['
	ArrayClose                = ']'
	DictionaryOpen  Delimiter = "<<"
	DictionaryClose Delimiter = ">>"
)

type Delimiter string

func (d Delimiter) Append(dst []byte) []byte {
	return append(dst, d...)
}
