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
	alignC uint8 = 0 + iota
	alignL
	alignR
	alignM = alignC
	alignT = alignL
	alignB = alignR
)

// Здесь подробно описано каждое поле настроек.
//
// Align -- позиционирование контента внутри объекта.
// Задается строкой, состоящей из символов 'L' (left), 'C' (center), 'R' (right) для позиционирования по горизонтали
// и символов 'T' (top), 'M' (middle), 'B' (bottom) для позиционирования по вертикали.
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
// Если не задано, используется шрифт, заданный методом Core.SetDefaultFont или Core.SetFontRegular. //TODO
//
// FontSize -- размер шрифта.
// Задается числом с плавающей запятой в пунктах unit.PT.
// Если не задано, используется размер, заданный по-умолчанию.
//
// TextColor -- цвет текста.
// Задается типом Color.
// Если не задано, по-умолчанию ColorBlack
//
// Wrap -- признак переноса текста.
// Задается булевым типом.
// Если не задано, по-умолчанию false.
// Текст с опцией Wrap автоматически перенесется на следующую строку, если его ширина будет превышать ширину ячейки.
// Высота строки (и всех ячеек в ней соответственно) будет увеличена под высоту области, которую занимает текст.
// Возможно задать принудительный перенос символом '\n'. Например:
//
//	.Cell("Универсальный\nпередаточный\nдокумент", CellOptions{ Wrap: true })
//
// Placeholder -- текст, используемый в случаях, когда текст ячейки пуст, или изображение не найдено.
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
type Options interface {
	CellOptions | NodeOptions | TableOptions | RowOptions | WatermarkOptions | PaginatorOptions
}

// NodeOptions используется для настройки параметров блока Block, слота Slot, заголовка Header.
// Подробное описание всех полей здесь: Options.
type NodeOptions struct {
	Border     string
	BorderSize unit.PT
	Spacing    unit.MM
	IndentH    unit.MM
	IndentV    unit.MM
	Ledge      unit.MM
}

// WatermarkOptions используется для настройки параметров водяного знака Watermark.
// Подробное описание всех полей здесь: Options.
type WatermarkOptions struct {
	Align string
}

// PaginatorOptions используется для настройки параметров пагинатора Paginator
// Подробное описание всех полей здесь: Options.
type PaginatorOptions struct {
	Offset int
	Align  string
}

// TableOptions используется для настройки параметров таблицы Table.
// Подробное описание всех полей здесь: Options.
type TableOptions struct {
	TextColor   Color
	BorderColor Color
	Border      string
	BorderSize  unit.PT
	IndentH     unit.MM
	IndentV     unit.MM
	SpacingH    unit.MM
	SpacingV    unit.MM
}

// RowOptions используется для настройки параметров строки Row.
// Подробное описание всех полей здесь: Options.
type RowOptions struct {
	MinHeight  unit.MM
	Border     string
	BorderSize unit.PT
}

// CellOptions используется для настройки параметров ячейки.
// Подробное описание всех полей здесь: Options.
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
	TextColor   Color
	BorderColor Color
	Wrap        bool
}

func getOptions[T Options](opts []T) T {
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
