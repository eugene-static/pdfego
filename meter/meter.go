package meter

import (
	"math"

	"golang.org/x/image/math/fixed"
)

const (
	dpi  = 72.0
	inch = 25.4
)

type MM float64

type PT float64

func (pt PT) MM() MM {
	return MM(pt / (dpi / inch))
}

func (pt PT) Float64() float64 {
	return float64(pt)
}

func (pt PT) FixedI() fixed.Int26_6 {
	return fixed.Int26_6(math.Round(pt.Float64() * 64.0))
}

func (mm MM) PT() PT {
	return PT(mm * (dpi / inch))
}

func (mm MM) Abs() (res MM) {
	res = mm

	if res < 0 {
		res = -res
	}

	return res
}

func FontHeight(fontSize PT) MM {
	return fontSize.MM() * 1 //TODO:
}
