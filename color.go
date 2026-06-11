package pdfego

// Color применяется для обозначения цвета текста и границ.
type Color uint32

// Базовая палитра (черный/белый/серый)
const (
	ColorBlack  Color = 0x000000 // Черный
	ColorWhite  Color = 0xFFFFFF // Белый
	ColorGray05 Color = 0x0D0D0D
	ColorGray10 Color = 0x1A1A1A
	ColorGray20 Color = 0x333333
	ColorGray30 Color = 0x4D4D4D
	ColorGray40 Color = 0x666666
	ColorGray50 Color = 0x808080
	ColorGray60 Color = 0x999999
	ColorGray70 Color = 0xB3B3B3
	ColorGray80 Color = 0xCCCCCC
	ColorGray90 Color = 0xE6E6E6
	ColorGray95 Color = 0xF2F2F2
)

// Основные цвета
const (
	ColorRed     Color = 0xFF0000 // Красный
	ColorGreen   Color = 0x00FF00 // Зеленый
	ColorBlue    Color = 0x0000FF // Синий
	ColorCyan    Color = 0x00FFFF // Светло-голубой
	ColorMagenta Color = 0xFF00FF // Маджента
	ColorYellow  Color = 0xFFFF00 // Желтый
)

// Приглушенные оттенки
const (
	ColorDarkRed   Color = 0x8B0000 // Темно-красный
	ColorCrimson   Color = 0xDC143C // Малиновый
	ColorFireBrick Color = 0xB22222 // Кирпичный
	ColorIndianRed Color = 0xCD5C5C // Индийский красный

	ColorDarkGreen   Color = 0x006400 // Темно-зеленый
	ColorForestGreen Color = 0x228B22 // Лесной зеленый
	ColorSeaGreen    Color = 0x2E8B57 // Морской зеленый
	ColorOlive       Color = 0x808000 // Оливковый

	ColorDarkBlue  Color = 0x00008B // Темно-синий
	ColorNavy      Color = 0x000080 // Морской
	ColorRoyalBlue Color = 0x4169E1 // Королевский синий
	ColorSteelBlue Color = 0x4682B4 // Стальной синий
	ColorLightBlue Color = 0xADD8E6 // Светло-голубой

	ColorDarkOrange Color = 0xFF8C00 // Темно-оранжевый
	ColorOrangeRed  Color = 0xFF4500 // Оранжево-красный
	ColorGold       Color = 0xFFD700 // Золотой
	ColorDarkGolden Color = 0xB8860B // Темно-золотой

	ColorPurple     Color = 0x800080 // Пурпурный
	ColorIndigo     Color = 0x4B0082 // Индиго
	ColorDarkViolet Color = 0x9400D3 // Темно-фиолетовый

	ColorBrown       Color = 0xA52A2A // Коричневый
	ColorSaddleBrown Color = 0x8B4513 // Седельно-коричневый
	ColorSienna      Color = 0xA0522D // Сиена
	ColorChocolate   Color = 0xD2691E // Шоколадный

	ColorTeal       Color = 0x008080 // Бирюзовый
	ColorTurquoise  Color = 0x40E0D0 // Бирюзовый светлый
	ColorAquamarine Color = 0x7FFFD4 // Аквамарин
)

const (
	ColorDefault = ColorBlack
)

// NewColor упаковывает новый цвет Color
//
//	red, green, blue в диапазоне от 0 до 255.
func NewColor(r, g, b uint8) Color {
	return (Color(r) << 16) | (Color(g) << 8) | Color(b)
}

func (c Color) rgb() (r float64, g float64, b float64) {
	r = float64(c >> 16 & 0xFF)
	g = float64(c >> 8 & 0xFF)
	b = float64(c & 0xFF)

	return r / 255, g / 255, b / 255
}

func (c Color) equal(other Color) bool {
	return c == other
}

func (c Color) isDefault() bool {
	return c == ColorDefault
}
