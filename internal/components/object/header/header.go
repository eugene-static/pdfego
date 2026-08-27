package header

import (
	"github.com/eugene-static/pdfego/internal/components/stream"
	"github.com/eugene-static/pdfego/internal/components/stream/primitives"
)

const (
	ver = "PDF-1.6"
)

type Header struct {
	Version      primitives.Comment
	BinaryMarker primitives.Comment
}

func New() *Header {
	return &Header{
		Version:      ver,
		BinaryMarker: primitives.BinaryMarker,
	}
}

func (h *Header) WriteToStream(str *stream.Stream) {
	str.NewStreamWriter().
		Write(h.Version).LF().
		Write(h.BinaryMarker).LF().
		Close()
}
