package font

type Color struct {
	r, g, b byte
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

func (c *Color) RGB() (r float64, g float64, b float64) {
	return float64(c.r) / 255, float64(c.g) / 255, float64(c.b) / 255
}

func (c *Color) IsBlack() bool {
	return c.r == 0 && c.g == 0 && c.b == 0
}
