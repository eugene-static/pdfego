package font

import (
	"github.com/eugene-static/pdf-craft/pkg/meter"
)

type Text struct {
	data  string
	width meter.MM
}

func (s *Text) Data() string {
	return s.data
}

func (s *Text) Width() meter.MM {
	return s.width
}
