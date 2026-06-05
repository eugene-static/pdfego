package pdf_craft

import (
	"github.com/eugene-static/pdf-craft/internal/buffer"
	"github.com/eugene-static/pdf-craft/pkg/unit"
)

const (
	borderLeft   uint8                                                 = 1 << iota // 0000 0001 (1)
	borderRight                                                                    // 0000 0010 (2)
	borderTop                                                                      // 0000 0100 (4)
	borderBottom                                                                   // 0000 1000 (8)
	thickLeft                                                                      // 0001 0000 (16)
	thickRight                                                                     // 0010 0000 (32)
	thickTop                                                                       // 0100 0000 (64)
	thickBottom                                                                    // 1000 0000 (128)
	borderAll    = borderLeft | borderRight | borderTop | borderBottom             // 0000 1111 (15)
	thickAll     = thickLeft | thickRight | thickTop | thickBottom                 // 1111 0000 (240)
)

const (
	alignL uint8 = 0 + iota
	alignC
	alignR
	alignT = alignL
	alignM = alignC
	alignB = alignR
)

type Color struct {
	r, g, b uint8
}

var (
	Black   = Color{0, 0, 0}
	White   = Color{255, 255, 255}
	Red     = Color{255, 0, 0}
	Green   = Color{0, 255, 0}
	Blue    = Color{0, 0, 255}
	Yellow  = Color{0, 255, 255}
	Cyan    = Color{255, 0, 255}
	Magenta = Color{255, 255, 0}
)

// NewColor инициализирует новый цвет Color
//
//	red, green, blue в диапазоне от 0 до 255.
func NewColor(r, g, b uint8) Color {
	return Color{r, g, b}
}

func NewColorFromHex(hex string) (Color, error) {
	return Black, nil
}

func (c *Color) RGB() (r float64, g float64, b float64) {
	return float64(c.r) / 255, float64(c.g) / 255, float64(c.b) / 255
}

func (c *Color) Equal(other Color) bool {
	return c.r == other.r && c.g == other.g && c.b == other.b
}

func (c *Color) isBlack() bool {
	return c.r == 0 && c.g == 0 && c.b == 0
}

type NodeOptions struct {
	Border     string
	BorderSize unit.PT
	Spacing    unit.MM
	IndentH    unit.MM
	IndentV    unit.MM
	Ledge      unit.MM
}

type WatermarkOptions struct {
	Align string
}

type TableOptions struct {
	Color      Color
	Border     string
	BorderSize unit.PT
	IndentH    unit.MM
	IndentV    unit.MM
	SpacingH   unit.MM
	SpacingV   unit.MM
}

type RowOptions struct {
	Height unit.MM
}

type CellOptions struct {
	ID          uint8
	Colspan     uint8
	Rowspan     uint8
	Align       string
	Border      string
	Font        string
	PlaceHolder string
	Scale       float64
	OffsetH     unit.MM
	OffsetV     unit.MM
	Height      unit.MM
	BorderSize  unit.PT
	FontSize    unit.PT
	Color       Color
	Wrap        bool
}

func getOptions[T CellOptions | NodeOptions | TableOptions | RowOptions | WatermarkOptions](opts []T) T {
	var opt T
	if len(opts) > 0 {
		opt = opts[0]
	}

	return opt
}

func renderBorder(buf *buffer.Buffer, x, y, w, h unit.MM, borderMask uint8, borderSize unit.PT) {
	var x0, y0, x1, y1 unit.MM

	if borderMask&borderAll == borderAll && (borderMask^borderAll == thickAll || borderMask^borderAll == 0) {
		bs := borderSize
		if borderMask&thickAll == thickAll {
			bs *= 3
		}

		buf.WriteRect(bs, x, y, w, h)

		return
	}

	if borderMask&borderLeft != 0 {
		bs := borderSize
		if borderMask&thickLeft != 0 {
			bs *= 3
		}

		x0 = x
		y0 = y
		x1 = x0
		y1 = y + h

		buf.WriteLine(bs, x0, y0, x1, y1)
	}

	if borderMask&borderRight != 0 {
		bs := borderSize
		if borderMask&thickRight != 0 {
			bs *= 3
		}

		x0 = x + w
		y0 = y
		x1 = x0
		y1 = y + h

		buf.WriteLine(bs, x0, y0, x1, y1)
	}

	if borderMask&borderTop != 0 {
		bs := borderSize
		if borderMask&thickTop != 0 {
			bs *= 3
		}

		x0 = x
		y0 = y
		x1 = x + w
		y1 = y

		buf.WriteLine(bs, x0, y0, x1, y1)
	}

	if borderMask&borderBottom != 0 {
		bs := borderSize
		if borderMask&thickBottom != 0 {
			bs *= 3
		}

		x0 = x
		y0 = y + h
		x1 = x + w
		y1 = y0

		buf.WriteLine(bs, x0, y0, x1, y1)
	}
}

func parseBorder(border string) (borderMask uint8) {
	for _, b := range border {
		switch b {
		case 'o':
			borderMask |= borderAll
		case 'O':
			borderMask |= borderAll | thickAll
		case 'l':
			borderMask |= borderLeft
		case 'L':
			borderMask |= borderLeft | thickLeft
		case 'r':
			borderMask |= borderRight
		case 'R':
			borderMask |= borderRight | thickRight
		case 't':
			borderMask |= borderTop
		case 'T':
			borderMask |= borderTop | thickTop
		case 'b':
			borderMask |= borderBottom
		case 'B':
			borderMask |= borderBottom | thickBottom
		}
	}

	return borderMask
}

func parseAlignment(alignment string) (alignH, alignV uint8) {
	alignH = alignC
	alignV = alignM

	for _, alg := range alignment {
		switch alg {
		case 'L':
			alignH = alignL
		case 'C':
			alignH = alignC
		case 'R':
			alignH = alignR
		case 'T':
			alignV = alignT
		case 'M':
			alignV = alignM
		case 'B':
			alignV = alignB
		default:
		}
	}

	return alignH, alignV
}

func coalesce[T comparable](vals ...T) T {
	var zero T

	for i := range vals {
		if vals[i] != zero {
			return vals[i]
		}
	}

	return zero
}
