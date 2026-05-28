package font

import "github.com/eugene-static/pdf-craft/meter"

type Segment struct {
	text  string
	width meter.MM
}

func (s *Segment) Text() string {
	return s.text
}

func (s *Segment) Width() meter.MM {
	return s.width
}
