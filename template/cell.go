package template

import (
	"strings"
	"unicode"
)

type cell struct {
	core       *Core
	x          float64
	y          float64
	width      float64
	height     float64
	colspan    int
	rowspan    int
	font       string
	fontSize   float64
	border     string
	borderSize float64
	align      string
	text       []string
	busy       bool
}

type CellOpts struct {
	Height     float64
	Colspan    int
	Rowspan    int
	Align      string
	Border     string
	BorderSize float64
	Font       string
	FontSize   byte
	Wrap       bool
}

// BT /[FontAlias] [FontSize] Tf [X] [Y] Td <[TextHex]> Tj ET
func (c *cell) render(buf *buffer) {
	if len(c.text) == 0 {
		return
	}

	f := c.core.getFont(c.font)

	dy := c.textDy()

	buf.print("BT ")
	buf.printFont(c.font, c.fontSize)

	for _, line := range c.text {
		dx := c.textDx(line)

		buf.printXY(c.x+dx, c.y+dy)
		buf.print(" Td ")
		buf.printText(f.face, line)
		buf.space()
	}

	buf.print("ET")

	if c.border != "" {
		c.drawBorder(buf)
	}
}

func (c *cell) drawBorder(buf *buffer) {
	for _, b := range c.border {
		if c.borderSize == 0 && unicode.IsUpper(b) {
			c.borderSize = c.core.border.thick
		}

		var x0, y0, x1, y1 float64

		switch b {
		case 'o', 'O':
			buf.printRect(c.borderSize, c.x, c.y, c.width, c.height)

			return
		case 't', 'T':
			x0 = c.x
			y0 = c.y
			x1 = c.x + c.width
			y1 = c.y
		case 'r', 'R':
			x0 = c.x + c.width
			y0 = c.y
			x1 = x0
			y1 = c.y + c.height
		case 'b', 'B':
			x0 = c.x
			y0 = c.y + c.height
			x1 = c.x + c.width
			y1 = y0
		case 'l', 'L':
			x0 = c.x
			y0 = c.y
			x1 = x0
			y1 = c.y + c.height
		default:
		}

		buf.printLine(c.borderSize, x0, y0, x1, y1)
	}
}

func (c *cell) textDx(text string) (dx float64) {
	switch {
	case strings.ContainsRune(c.align, 'R'):
		dx = c.width - c.core.measureText(c.font, c.fontSize, text)
	case strings.ContainsRune(c.align, 'C'):
		dx = (c.width - c.core.measureText(c.font, c.fontSize, text)) / 2
	default:
		dx = 0
	}

	return dx
}

func (c *cell) textDy() (dy float64) {
	lenLines := len(c.text)

	switch {
	case strings.ContainsRune(c.align, 'B'):
		dy = c.height - float64(lenLines)*c.core.fontHeight
	case strings.ContainsRune(c.align, 'M'):
		dy = (c.height - float64(lenLines)*c.core.fontHeight) / 2
	default:
		dy = 0
	}

	return dy
}
