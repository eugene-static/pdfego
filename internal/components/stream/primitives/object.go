package primitives

import (
	"github.com/eugene-static/pdfego/internal/components/stream"
)

type XObject struct {
	name   Alias
	matrix Matrix
}

func NewXObject(name Alias, matrix Matrix) *XObject {
	return &XObject{
		name:   name,
		matrix: matrix,
	}
}

// q $W 0 0 $H $X $Y cm /$Name Do Q
func (obj *XObject) WriteToStream(dst *stream.Stream) {
	dst.NewStreamWriter().
		Write(SaveGraphicsState).SP().
		Write(obj.matrix).SP().
		Write(ConcatenateMatrix).SP().
		Write(obj.name).SP().
		Write(InvokeNamedXObject).SP().
		Write(RestoreGraphicsState).LF().
		Close()
}
