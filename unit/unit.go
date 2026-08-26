package unit

import (
	"math"
	"strconv"

	"github.com/eugene-static/pdfego/internal/components/stream/bytes"
	"golang.org/x/image/math/fixed"
)

const (
	dpi  = 72.0
	inch = 25.4
)

type PT float64

func (pt PT) MM() MM {
	return MM(pt / (dpi / inch))
}

func (pt PT) Float64() float64 {
	return float64(pt)
}

func (pt PT) Add(pt2 PT) PT {
	return pt + pt2
}

func (pt PT) FixedI() fixed.Int26_6 {
	return fixed.Int26_6(math.Round(pt.Float64() * 64.0))
}

func (pt PT) Sub(pt2 PT) PT {
	return pt - pt2
}

func (pt PT) Abs() PT {
	if pt < 0 {
		return -pt
	}

	return pt
}

func (pt PT) Append(dst []byte) []byte {
	if pt == PT(int(pt)) {
		return strconv.AppendInt(dst, int64(pt), 10)
	}

	return bytes.AppendFloat(dst, float64(pt))
}

type MM float64

func (mm MM) PT() PT {
	return PT(mm * (dpi / inch))
}

func (mm MM) Float64() float64 {
	return float64(mm)
}

func (mm MM) Add(mm2 MM) MM {
	return mm + mm2
}

func (mm MM) Abs() (res MM) {
	res = mm

	if res < 0 {
		res = -res
	}

	return res
}

func (mm MM) Neg() MM {
	return -mm
}

func (mm MM) Append(dst []byte) []byte {
	if mm == MM(int(mm)) {
		return strconv.AppendInt(dst, int64(mm), 10)
	}

	return bytes.AppendFloat(dst, float64(mm))
}

type Intensity float64

func (in Intensity) Append(dst []byte) []byte {
	return bytes.AppendFloat(dst, float64(in))
}

type Point struct {
	x MM
	y MM
}

func NewPoint(x MM, y MM) Point {
	return Point{x: x, y: y}
}

func (pt *Point) X() MM {
	return pt.x
}

func (pt *Point) Y() MM {
	return pt.y
}

func (pt *Point) SetX(x MM) {
	pt.x = x
}

func (pt *Point) SetY(y MM) {
	pt.y = y
}

func (pt *Point) AddX(x MM) {
	pt.x += x
}

func (pt *Point) AddY(y MM) {
	pt.y += y
}
