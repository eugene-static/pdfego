package primitives

const (
	Object    Keyword = "obj"
	EndObject Keyword = "endobj"
	Stream    Keyword = "stream"
	EndStream Keyword = "endstream"

	BeginBfChar Keyword = "beginbfchar"
	EndBfChar   Keyword = "endbfchar"
	EndCMap     Keyword = "endcmap"

	XRef      Keyword = "xref"
	Trailer   Keyword = "trailer"
	StartXRef Keyword = "startxref"

	True  Keyword = "true"
	False Keyword = "false"
	Null  Keyword = "null"

	Reference     = 'R'
	NamePrefix    = '/'
	CommentPrefix = '%'
	RecordIsUsed  = 'n'
	RecordNotUsed = 'f'

	BinaryMarker Comment = "\x80\x80\x80\x80"
	EOF          Comment = "%EOF"
)

type Keyword string

func (kw Keyword) Append(dst []byte) []byte {
	return append(dst, kw...)
}

func (kw Keyword) String() string {
	return string(kw)
}

type Alias string

func (alias Alias) Append(dst []byte) []byte {
	dst = append(dst, NamePrefix)

	return append(dst, alias...)
}

type Comment string

func (comm Comment) Append(dst []byte) []byte {
	dst = append(dst, CommentPrefix)

	return append(dst, comm...)
}

func (comm Comment) String() string {
	comment := make([]byte, 0, len(comm)+1)

	comment = comm.Append(comment)

	return string(comment)
}
