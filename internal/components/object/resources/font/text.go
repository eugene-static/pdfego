package font

import (
	"github.com/eugene-static/pdfego/internal/components/stream/primitives"
	"github.com/eugene-static/pdfego/unit"
)

type Text struct {
	symbols []Symbol
	matrix  primitives.Matrix
	width   unit.MM
	shift   unit.EM
}

func (text *Text) SetMatrix(point primitives.Point) {
	matrix := primitives.NewDefaultTextMatrix(point)

	text.matrix = matrix
}

func (text *Text) Width() unit.MM {
	return text.width
}

type Symbol struct {
	hex     unit.HEX
	fontID  uint8
	isSpace bool
}
