package template

import (
	"strings"
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
func (c *cell) render() {
	//if f.y+h > f.pageBreakTrigger && !f.inHeader && !f.inFooter && f.acceptPageBreak() {
	//	// Automatic page break
	//	x := f.x
	//	ws := f.ws
	//	// dbg("auto page break, x %.2f, ws %.2f", x, ws)
	//	if ws > 0 {
	//		f.ws = 0
	//		f.out("0 Tw")
	//	}
	//	f.AddPageFormat(f.curOrientation, f.curPageSize)
	//	if f.err != nil {
	//		return
	//	}

	if len(c.text) == 0 {
		return
	}

	dy := c.textDy()

	c.core.print("BT ")
	c.core.printFont(c.font, c.fontSize)

	for _, line := range c.text {
		dx := c.textDx(line)
		x, y := c.core.xy()

		c.core.printXY(x+dx, y+dy)
		c.core.print(" Td ")
		c.core.printText(c.font, line)
		c.core.printSpace()
	}

	c.core.print("ET")

	if c.border != "" {
		c.drawBorder()
	}
}

func (c *cell) drawBorder() {
	x, y := c.core.xy()

	if strings.ContainsRune(c.border, '+') {
		c.borderSize = c.core.border.thick
	}

	if c.border == "1" {
		// 1 w x0 y0 w h re S
		c.core.printFloat64(c.borderSize)
		c.core.print(" w ")
		c.core.printXY(x, y)
		c.core.printSpace()
		c.core.printXY(c.width, -c.height)
		c.core.print(" re S ")

		return
	}

	if strings.ContainsRune(c.border, 'T') {
		// Top: верхняя граница
		x0 := x
		y0 := y
		x1 := x + c.width
		y1 := y

		c.core.printLine(c.borderSize, x0, y0, x1, y1)
	}

	if strings.ContainsRune(c.border, 'R') {
		// Right: правая граница
		x0 := x + c.width
		y0 := y
		x1 := x0
		y1 := y - c.height

		c.core.printLine(c.borderSize, x0, y0, x1, y1)
	}

	if strings.ContainsRune(c.border, 'B') {
		// Bottom: нижняя граница
		x0 := x
		y0 := y - c.height
		x1 := x + c.width
		y1 := y0

		c.core.printLine(c.borderSize, x0, y0, x1, y1)
	}

	if strings.ContainsRune(c.border, 'L') {
		// Left: левая граница
		x0 := x
		y0 := y
		x1 := x0
		y1 := y - c.height

		c.core.printLine(c.borderSize, x0, y0, x1, y1)
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

	return -dy
}
