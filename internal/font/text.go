package font

import (
	"github.com/eugene-static/pdf-craft/pkg/unit"
)

type Text struct {
	data  string
	width unit.MM
}

func (s *Text) Data() string {
	return s.data
}

func (s *Text) Width() unit.MM {
	return s.width
}
