package primitives

import (
	"github.com/eugene-static/pdfego/internal/components/stream/bytes"
	"github.com/eugene-static/pdfego/unit"
)

type Matrix struct {
	scale Scale
	skew  Skew
	point Point
}

func NewMatrix(scale Scale, skew Skew, point Point) Matrix {
	return Matrix{
		scale: scale,
		skew:  skew,
		point: point,
	}
}

func NewDefaultTextMatrix(point Point) Matrix {
	return Matrix{
		scale: NewScale(1, 1),
		point: point,
	}
}

func (m Matrix) Append(dst []byte) []byte {
	return bytes.NewWriter(dst).
		Write(m.scale.x).SP().
		Write(m.skew.x).SP().
		Write(m.skew.y).SP().
		Write(m.scale.y).SP().
		Write(m.point).Bytes()
}

type Point struct {
	x unit.PT
	y unit.PT
}

func NewPoint(x, y unit.PT) Point {
	return Point{
		x: x,
		y: y,
	}
}

func (p Point) Append(dst []byte) []byte {
	return bytes.NewWriter(dst).
		Write(p.x).SP().
		Write(p.y).Bytes()
}

type Size struct {
	width  unit.PT
	height unit.PT
}

func NewSize(width, height unit.PT) Size {
	return Size{
		width:  width,
		height: height,
	}
}

func (s Size) Append(dst []byte) []byte {
	return bytes.NewWriter(dst).
		Write(s.width).SP().
		Write(s.height).Bytes()
}

type Scale struct {
	x unit.PT
	y unit.PT
}

func NewScale(x, y unit.PT) Scale {
	return Scale{
		x: x,
		y: y,
	}
}

func (s Scale) Append(dst []byte) []byte {
	return bytes.NewWriter(dst).
		Write(s.x).SP().
		Write(s.y).Bytes()
}

type Skew struct {
	x unit.PT
	y unit.PT
}

func NewSkew(x, y unit.PT) Skew {
	return Skew{
		x: x,
		y: y,
	}
}

func (s Skew) Append(dst []byte) []byte {
	return bytes.NewWriter(dst).
		Write(s.x).SP().
		Write(s.y).Bytes()
}
