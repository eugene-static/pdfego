package template

import (
	"log/slog"
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
	fontSize   int
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
	FontSize   int
	Wrap       bool
}

// BT /[FontAlias] [FontSize] Tf 1 0 0 1 [X] [Y] Tm <[TextHex]> Tj ET
func (c *cell) render(buf *buffer) {
	if len(c.text) == 0 {
		return
	}

	f := c.core.getFont(c.font)

	buf.print("BT\n")
	buf.printFont(c.font, c.fontSize)

	for i, line := range c.text {
		dx := c.textDx(line)
		dy := c.textDy(i)

		x, y := c.core.ptXY(c.x+dx, c.y+dy)

		// "1 0 0 1 x y Tm" задает абсолютную позицию текста на странице.
		buf.print("1 0 0 1 ")
		buf.printXY(x, y)
		buf.print(" Tm ")
		buf.printText(f, line)
		buf.print(" Tj\n")
	}

	buf.print("ET\n")

	if c.border != "" {
		c.drawBorder(buf)
	}
}

func (c *cell) drawBorder(buf *buffer) {
	for _, b := range c.border {
		bs := c.borderSize

		if c.borderSize == 0 {
			bs = c.core.border.thin
			if unicode.IsUpper(b) {
				bs = c.core.border.thick
			}
		}

		var x0, y0, x1, y1 float64

		switch b {
		case 'o', 'O':
			x, y := c.core.ptXY(c.x, c.y)
			w, h := pt(c.width), pt(c.height)

			buf.printRect(bs, x, y, w, h)

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
			continue
		}

		x0, y0 = c.core.ptXY(x0, y0)
		x1, y1 = c.core.ptXY(x1, y1)

		buf.printLine(bs, x0, y0, x1, y1)
	}
}

func (c *cell) textDx(text string) (dx float64) {
	f := c.core.getFont(c.font)

	switch {
	case strings.ContainsRune(c.align, 'R'):
		textWidth := f.measureText(c.fontSize, text)

		c.core.log.Debug("width", slog.String("text", text), slog.Float64("mm", textWidth))

		dx = c.width - textWidth - 0.2
	case strings.ContainsRune(c.align, 'C'):
		textWidth := f.measureText(c.fontSize, text)

		c.core.log.Debug("width", slog.String("text", text), slog.Float64("mm", textWidth))

		dx = (c.width - textWidth) / 2
	default:
		// чтобы текст не прилипал к границе
		dx = 0.2
	}

	return dx
}

func (c *cell) textDy(index int) (dy float64) {
	lenLines := len(c.text)

	switch {
	case strings.ContainsRune(c.align, 'B'):
		dy = c.height - float64(lenLines)*c.core.fontHeight
	case strings.ContainsRune(c.align, 'M'):
		dy = (c.height - float64(lenLines)*c.core.fontHeight) / 2
	default:
		dy = 0.1
	}

	// index + 1 необходим для того, чтобы выставить Y-координату по верхнему краю шрифта.
	// PDF считает Y от нижней границы страницы, а здесь все координаты указаны от верхней. К тому же позиционирует шрифт по baseline.
	// Поэтому для верного позиционирования шрифта нам необходимо добавить еще одну высоту строки.
	dy += float64(index+1) * c.core.fontHeight

	return dy
}
