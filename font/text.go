package font

import "github.com/eugene-static/pdf-craft/meter"

type Segment struct {
	text   string
	width  meter.MM
	shifts []int
}
