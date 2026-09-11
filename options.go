package pdfego

import (
	"github.com/eugene-static/pdfego/internal/components/stream"
	"github.com/eugene-static/pdfego/internal/components/stream/primitives"
	"github.com/eugene-static/pdfego/unit"
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
	alignC uint8 = 0 + iota
	alignL
	alignR
	alignJ

	alignM = alignC
	alignT = alignL
	alignB = alignR
)

const (
	placementTop uint8 = iota
	placementBottom
)

const (
	fitOff uint8 = iota
	fitSymbols
	fitWords
)

// Здесь подробно описано каждое поле настроек.
//
// Align -- позиционирование контента внутри объекта.
// Задается строкой, состоящей из символов 'L' (left), 'C' (center), 'R' (right) для позиционирования по горизонтали.
// И символов 'T' (top), 'M' (middle), 'B' (bottom) для позиционирования по вертикали.
// 'J' (jusify) используется для растягивания текста по ширине.
// Если не задано, используется по-умолчанию "CM".
//
// Border -- отрисовка границ объекта.
// Задается строкой, состоящей из символов 'l' (left), 't' (top), 'r' (right), 'b' (bottom). 'o' (outline) используется для отрисовки всех границ.
// Если не задано, используется толщина заданная по-умолчанию.
// Символами в верхнем регистре ('L', 'T', 'R', 'B', 'O' соответственно) обозначается толстая линия (тонкая х3).
// Комбинировать символы можно в любом порядке.
// Толщина линии границ задается параметром BorderSize.
//
// BorderSize -- толщина тонких линий границ объекта.
// Задается числом с плавающей запятой в пунктах unit.PT.
// Если не задано, используется толщина, заданная по-умолчанию.
//
// BorderColor -- цвет линий границ объекта.
// Задается типом Color.
// Если не задано, по-умолчанию ColorBlack
//
// Font -- шрифт текста.
// Задается строкой с псевдонимом шрифта, инициализированного методом Core.ReadFont.
// Если не задано, используется шрифт, заданный методом Core.SetFontRegular.
//
// FontSize -- размер шрифта.
// Задается числом с плавающей запятой в пунктах unit.PT.
// Если не задано, используется размер, заданный по-умолчанию.
//
// TextColor -- цвет текста.
// Задается типом Color.
// Если не задано, по-умолчанию ColorBlack
//
// Wrap -- признак переноса текста. (Deprecated)
//
// FitContent -- способ заполнения ячейки.
// Задается строкой, где "W" (words) означает перенос текста по словам, "S" означает перенос текста по символам.
// Если не задано, по-умолчанию тект не переносится.
// Текст с опцией FitContent автоматически перенесется на следующую строку, если его ширина будет превышать ширину ячейки.
// Высота строки (и всех ячеек в ней соответственно) будет увеличена под высоту области, которую занимает текст.
// Возможно задать принудительный перенос символом '\n'. Например:
//
//	.Cell("Универсальный\nпередаточный\nдокумент", CellOptions{ FitContent: "W" })
//
// Placeholder -- текст, используемый в случаях, когда изображение не найдено.
//
// Scale -- масштаб изображения.
// Задается числом с плавающей запятой.
// По-умолчанию 1.
//
// OffsetH -- сдвиг изображения по-горизонтали.
// Задается числом с плавающей запятой в миллиметрах unit.MM.
//
// OffsetV -- сдвиг изображения по-вертикали.
// Задается числом с плавающей запятой в миллиметрах unit.MM.
//
// Height -- высота.
// Задается числом с плавающей запятой в миллиметрах unit.MM.
//
// MinHeight -- минимальная высота.
// Задается числом с плавающей запятой в миллиметрах unit.MM.
//
// IndentH -- отступ объекта от родительского по-горизонтали.
// Задается числом с плавающей запятой в миллиметрах unit.MM.
//
// IndentV -- отступ объекта от родительского по-вертикали.
// Задается числом с плавающей запятой в миллиметрах unit.MM.
//
// Ledge -- выступ объекта по-вертикали.
// Задается числом с плавающей запятой в миллиметрах unit.MM.
// Является частью объекта, что влияет на его высоту и отрисовку границ.
//
// Spacing -- расстояние между дочерними элементами объекта.
// Задается числом с плавающей запятой в миллиметрах unit.MM.
//
// SpacingH -- расстояние между дочерними элементами объекта, выстроенными по-горизонтали.
// Задается числом с плавающей запятой в миллиметрах unit.MM.
//
// SpacingV -- расстояние между дочерними элементами объекта, выстроенными по-вертикали.
// Задается числом с плавающей запятой в миллиметрах unit.MM.
//
// Colspan -- количество колонок, которое занимает ячейка.
//
// Rowspan -- количество строк, которое занимает ячейка.
//
// ID -- идентификатор ячейки.
//
// Offset -- начальная точка отсчета.
//
// Skip -- количество страниц, которые необходимо пропустить, перед началом печати ее номера.
//
// Placement -- размещение блока на странице. В настоящее время применяется только для пагинации.
// Значения: "T", "B".
type Options interface {
	CellOptions | NodeOptions | TableOptions | RowOptions | WatermarkOptions | PaginatorOptions
}

// NodeOptions используется для настройки параметров блока Block, слота Slot, заголовка Header.
// Подробное описание всех полей здесь: Options.
type NodeOptions struct {
	Border        string
	BorderColor   Color
	BorderPattern []unit.PT
	BorderSize    unit.PT
	Spacing       unit.MM
	IndentH       unit.MM
	IndentV       unit.MM
	Ledge         unit.MM
}

// WatermarkOptions используется для настройки параметров водяного знака Watermark.
// Подробное описание всех полей здесь: Options.
type WatermarkOptions struct {
	Align string
}

// PaginatorOptions используется для настройки параметров пагинатора Paginator.
// Подробное описание всех полей здесь: Options.
type PaginatorOptions struct {
	Offset    int
	Skip      uint
	Align     string
	Placement string
}

// TableOptions используется для настройки параметров таблицы Table.
// Подробное описание всех полей здесь: Options.
type TableOptions struct {
	TextColor     Color
	BorderColor   Color
	Border        string
	BorderPattern []unit.PT
	BorderSize    unit.PT
	IndentH       unit.MM
	IndentV       unit.MM
	SpacingH      unit.MM
	SpacingV      unit.MM
}

// RowOptions используется для настройки параметров строки Row.
// Подробное описание всех полей здесь: Options.
type RowOptions struct {
	MinHeight unit.MM
}

// CellOptions используется для настройки параметров ячейки.
// Подробное описание всех полей здесь: Options.
type CellOptions struct {
	Align         string
	Border        string
	Font          string
	PlaceHolder   string
	FitContent    string
	Scale         float64
	OffsetH       unit.MM
	OffsetV       unit.MM
	Height        unit.MM
	BorderPattern []unit.PT
	BorderSize    unit.PT
	FontSize      unit.PT
	TextColor     Color
	BorderColor   Color
	ID            uint8 // Deprecated
	Colspan       uint8
	Rowspan       uint8
	UnderLine     bool
	Wrap          bool // Deprecated
}

type borderOptions struct {
	mask        uint8
	color       Color
	parentColor Color
	size        unit.PT
	pattern     []unit.PT
}

func BorderPattern(points ...unit.PT) []unit.PT {
	return points
}

var BorderPatternDash = []unit.PT{2, 1}

type mode uint8

const (
	skipMode mode = iota
	textMode
	imageMode
	counterMode
)

const (
	orderMode mode = iota
	iteratorMode
)

func (m mode) is(m2 mode) bool {
	return m == m2
}

const (
	zeroIndex          uint8   = 255
	borderSizeModified unit.PT = 3
)

func getOptions[T Options](opts []T) T {
	var opt T

	if len(opts) > 0 {
		opt = opts[0]
	}

	return opt
}

func renderBorder(dst *stream.Stream, cursor unit.Point, w, h unit.MM, border borderOptions) {
	if border.color != border.parentColor {
		border.color.WriteToStream(dst, primitives.StrokeColorRGB)
		defer border.parentColor.WriteToStream(dst, primitives.StrokeColorRGB)
	}

	if len(border.pattern) > 0 {
		primitives.NewLinePattern(border.pattern).WriteToStream(dst)
		defer primitives.DefaultLinePattern().WriteToStream(dst)
	}

	if border.mask&borderAll == borderAll && (border.mask^borderAll == thickAll || border.mask^borderAll == 0) {
		bs := getBorderSize(border.size, border.mask, thickAll)
		start := primitives.NewPoint(cursor.X().PT(), cursor.Y().PT().Abs())
		size := primitives.NewSize(w.PT(), h.Neg().PT())

		primitives.
			NewRect(bs, start, size).
			WriteToStream(dst)

		return
	}

	if border.mask&borderLeft != 0 {
		bs := getBorderSize(border.size, border.mask, thickLeft)
		start := primitives.NewPoint(cursor.X().PT(), cursor.Y().PT().Abs())
		end := primitives.NewPoint(cursor.X().PT(), cursor.Y().Add(h).PT().Abs())

		primitives.
			NewLine(bs, start, end).
			WriteToStream(dst)
	}

	if border.mask&borderRight != 0 {
		bs := getBorderSize(border.size, border.mask, thickRight)
		start := primitives.NewPoint(cursor.X().Add(w).PT(), cursor.Y().PT().Abs())
		end := primitives.NewPoint(cursor.X().Add(w).PT(), cursor.Y().Add(h).PT().Abs())

		primitives.
			NewLine(bs, start, end).
			WriteToStream(dst)
	}

	if border.mask&borderTop != 0 {
		bs := getBorderSize(border.size, border.mask, thickTop)
		start := primitives.NewPoint(cursor.X().PT(), cursor.Y().PT().Abs())
		end := primitives.NewPoint(cursor.X().Add(w).PT(), cursor.Y().PT().Abs())

		primitives.
			NewLine(bs, start, end).
			WriteToStream(dst)
	}

	if border.mask&borderBottom != 0 {
		bs := getBorderSize(border.size, border.mask, thickBottom)
		start := primitives.NewPoint(cursor.X().PT(), cursor.Y().Add(h).PT().Abs())
		end := primitives.NewPoint(cursor.X().Add(w).PT(), cursor.Y().Add(h).PT().Abs())

		primitives.
			NewLine(bs, start, end).
			WriteToStream(dst)
	}
}

func getBorderSize(bs unit.PT, bMask, size uint8) unit.PT {
	if bMask&size != 0 {
		bs *= borderSizeModified
	}

	return bs
}

func parseBorder(border string) (borderMask uint8) {
	if len(border) > 4 {
		border = border[:4]
	}

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
	if len(alignment) > 2 {
		alignment = alignment[:2]
	}

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
		case 'J':
			alignH = alignJ
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

func parsePlacement(placement string) (plm uint8) {
	plm = placementBottom

	if len(placement) == 0 {
		return plm
	}

	switch placement[0] {
	case 'T':
		plm = placementTop
	case 'B':
		plm = placementBottom
	default:
	}

	return plm
}

func parseFitContent(wrap bool, fitContent string) (fit uint8) {
	fit = fitOff

	if fitContent != "" {
		switch fitContent[0] {
		case 'W':
			fit = fitWords
		case 'S':
			fit = fitSymbols
		default:
			fit = fitOff
		}
	}

	if wrap {
		fit = fitWords
	}

	return fit
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
