package primitives

import (
	"github.com/eugene-static/pdfego/internal/components/stream"
	"github.com/eugene-static/pdfego/unit"
)

type Rect struct {
	strokeWidth unit.PT
	start       Point
	size        Size
}

func NewRect(strokeWidth unit.PT, start Point, size Size) *Rect {
	return &Rect{
		strokeWidth: strokeWidth,
		start:       start,
		size:        size,
	}
}

// $Size w $X0 $Y0 m $X1 $Y1 l S
func (r *Rect) WriteToStream(dst *stream.Stream) {
	dst.NewStreamWriter().
		Write(r.strokeWidth).SP().
		Write(Width).SP().
		Write(r.start).SP().
		Write(r.size).SP().
		Write(Rectangle).SP().
		Write(Stroke).LF().
		Close()
}

type Line struct {
	strokeWidth unit.PT
	start       Point
	end         Point
}

func NewLine(strokeWidth unit.PT, start, end Point) *Line {
	return &Line{
		strokeWidth: strokeWidth,
		start:       start,
		end:         end,
	}
}

// $Size w $X0 $Y0 m $X1 $Y1 l S
func (l *Line) WriteToStream(dst *stream.Stream) {
	dst.NewStreamWriter().
		Write(l.strokeWidth).SP().
		Write(Width).SP().
		Write(l.start).SP().
		Write(MoveTo).SP().
		Write(l.end).SP().
		Write(LineTo).SP().
		Write(Stroke).LF().
		Close()
}
