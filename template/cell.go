package template

import (
	"strings"
	"unicode"

	"github.com/eugene-static/pdf-craft/buffer"
	"github.com/eugene-static/pdf-craft/core"
	"github.com/eugene-static/pdf-craft/font"
	"github.com/eugene-static/pdf-craft/meter"
)

type cell struct {
	core       *core.Core
	x          meter.MM
	y          meter.MM
	width      meter.MM
	height     meter.MM
	colspan    int
	rowspan    int
	font       string
	fontSize   meter.PT
	border     string
	borderSize meter.PT
	align      string
	text       []segment
	busy       bool
}

type segment struct {
	text  string
	width meter.MM
	shift []int
}

type CellOpts struct {
	Height     float64
	Colspan    int
	Rowspan    int
	Align      string
	Border     string
	BorderSize float64
	Font       string
	FontSize   meter.PT
	Wrap       bool
}

// BT /[FontAlias] [FontSize] Tf 1 0 0 1 [X] [Y] Tm <[TextHex]> Tj ET
func (c *cell) render(buf *buffer.Buffer) {
	if len(c.text) == 0 {
		return
	}

	f := c.core.Font(c.font)

	buf.WriteStringLn("BT")
	buf.WriteFont(c.font, c.fontSize)

	for i, seg := range c.text {
		dx := c.textDx(f, seg.text)
		dy := c.textDy(i)

		x := c.x + dx
		y := c.y + dy

		buf.WriteText(f, x, y, seg.text, seg.shift)
	}

	buf.WriteStringLn("ET")

	if c.border != "" {
		c.renderBorder(buf)
	}
}

func (c *cell) renderBorder(buf *buffer.Buffer) {
	bs := c.borderSize

	for _, b := range c.border {
		if unicode.IsUpper(b) {
			bs = c.core.BorderThick()
		}

		var x0, y0, x1, y1 meter.MM

		switch b {
		case 'o', 'O':
			buf.WriteRect(bs, c.x, c.y, c.width, c.height)

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

		buf.WriteLine(bs, x0, y0, x1, y1)
	}
}

func (c *cell) textDx(f *font.Font, text string) (dx meter.MM) {
	switch {
	case strings.ContainsRune(c.align, 'R'):
		textWidth := f.MeasureText(c.fontSize, text)

		dx = c.width - textWidth.MM() - 0.2
	case strings.ContainsRune(c.align, 'C'):
		textWidth := f.MeasureText(c.fontSize, text)

		dx = (c.width - textWidth.MM()) / 2
	default:
		// чтобы текст не прилипал к границе
		dx = 0.2
	}

	return dx
}

func (c *cell) textDy(index int) (dy meter.MM) {
	lenLines := meter.MM(len(c.text))
	fontHeight := c.fontSize.MM()

	k := meter.MM(0.2)

	// index + 1 необходим для того, чтобы выставить Y-координату по верхнему краю шрифта.
	// PDF считает Y от нижней границы страницы, а здесь все координаты указаны от верхней. К тому же позиционирует шрифт по baseline.
	// Поэтому для верного позиционирования шрифта нам необходимо добавить еще одну высоту строки.
	baseLineDy := meter.MM(index+1) * fontHeight

	switch {
	case strings.ContainsRune(c.align, 'B'):
		dy = c.height - lenLines*fontHeight - k
	case strings.ContainsRune(c.align, 'M'):
		dy = (c.height - lenLines*fontHeight) / 2
	default:
		dy = k
	}

	dy += baseLineDy

	return dy
}
