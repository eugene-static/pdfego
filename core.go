package pdfego

import (
	"errors"
	"os"

	"github.com/eugene-static/pdfego/internal/components/object/pages"
	"github.com/eugene-static/pdfego/internal/components/object/resources"
	"github.com/eugene-static/pdfego/internal/components/object/resources/font"
	"github.com/eugene-static/pdfego/internal/components/object/resources/image"
	"github.com/eugene-static/pdfego/internal/compressor"
	"github.com/eugene-static/pdfego/unit"
)

const (
	FontRegular = "REG"
	FontBold    = "BOLD"
	FontItalic  = "ITALIC"

	Portrait  = "P"
	Landscape = "L"
)

// Core -- ядро конструктора. Позволяет настроить конструктор единожды и переиспользовать эти настройки при каждой новой генерации.
// Так же содержит информацию, необходимую для различных узлов конструктора.
type Core struct {
	compressor          *compressor.Compressor
	resources           *resources.Resources
	pagesConfig         *pages.Config
	fontRegularAlias    string
	fontBoldAlias       string
	fontItalicAlias     string
	fontSize            unit.PT
	borderSize          unit.PT
	ignoreImageNotFound bool
	enableCompression   bool
}

// NewCore инициализирует новый экземпляр ядра конструктора с заранее заданной ориентацией страницы.
// "P" -- портретная ориентация (по умолчанию), "L" -- альбомная. Границы соответствуют формату А4.
func NewCore(orientation string) *Core {
	width := pages.A4Width.MM()
	height := pages.A4Height.MM()

	if orientation == Landscape {
		width, height = height, width
	}

	return &Core{
		compressor:       compressor.New(),
		resources:        resources.New(),
		pagesConfig:      pages.NewConfig(width, height),
		fontRegularAlias: FontRegular,
		fontBoldAlias:    FontBold,
		fontItalicAlias:  FontItalic,
		fontSize:         unit.PT(6),
		borderSize:       unit.PT(0.3),
	}
}

// EnableCompression включает компрессию страниц документа. Значительно уменьшает объем файла, но увеличивает время на генерацию.
func (core *Core) EnableCompression() {
	core.enableCompression = true
}

func (core *Core) DisableCompression() {
	core.enableCompression = false
}

// IgnoreImageNotFound игнорирует ошибку, если изображение не найдено. Вместо ненайденного изображение будет использован Placeholder.
func (core *Core) IgnoreImageNotFound() {
	core.ignoreImageNotFound = true
}

// ReadFont добавляет новый шрифт с заданным псевдонимом alias, который потом можно использовать в каждой ячейке таблицы.
func (core *Core) ReadFont(path, alias string) error {
	if alias == "" {
		err := errors.New("псевдоним не может быть пустым")

		return err
	}

	f, err := font.New(path, alias, core.compressor)
	if err != nil {
		return err
	}

	core.resources.SetFont(alias, f)

	return nil
}

// SetFontRegular читает и устанавливает шрифт с псевдонимом "REG". При конфигурации ячейки указывать этот псевдоним не обязательно, если не задан шрифт по-умолчанию.
func (core *Core) SetFontRegular(path string) error {
	err := core.ReadFont(path, FontRegular)
	if err != nil {
		return err
	}

	return nil
}

// SetFontBold читает и устанавливает шрифт с псевдонимом "BOLD". При конфигурации ячейки с использованием метода LabelBold указывать этот псевдоним не обязательно.
func (core *Core) SetFontBold(path string) error {
	err := core.ReadFont(path, FontBold)
	if err != nil {
		return err
	}

	return nil
}

// SetFontItalic читает и устанавливает шрифт с псевдонимом "ITALIC".
func (core *Core) SetFontItalic(path string) error {
	err := core.ReadFont(path, FontItalic)
	if err != nil {
		return err
	}

	return nil
}

// SetDefaultFont устанавливает шрифт по-умолчанию. Если не задан, используется REG.
func (core *Core) SetDefaultFont(alias string) error {
	_, err := core.resources.GetFont(alias)
	if err != nil {
		return err
	}

	core.fontRegularAlias = alias

	return nil
}

// SetDefaultFontBold устанавливает жирный шрифт по-умолчанию. Если не задан, используется BOLD.
func (core *Core) SetDefaultFontBold(alias string) error {
	_, err := core.resources.GetFont(alias)
	if err != nil {
		return err
	}

	core.fontBoldAlias = alias

	return nil
}

// SetDefaultFontItalic устанавливает наклонный шрифт по-умолчанию. Если не задан, используется ITALIC.
func (core *Core) SetDefaultFontItalic(alias string) error {
	_, err := core.resources.GetFont(alias)
	if err != nil {
		return err
	}

	core.fontBoldAlias = alias

	return nil
}

// SetDefaultFontSize устанавливает размер шрифта для документа в пунктах. При конфигурации ячейки без явного указания размера будет использоваться заданный этой функцией.
//
// По-умолчанию: 6pt.
func (core *Core) SetDefaultFontSize(fontSize unit.PT) {
	core.fontSize = fontSize
}

// SetDefaultBorderSize устанавливает размер тонкой линии для документа в пунктах. При конфигурации ячейки без явного указания размера будет использоваться заданный этой функцией.
// Для толстых линий будет использован это значение х3.
//
// По-умолчанию: 0.3pt.
func (core *Core) SetDefaultBorderSize(size unit.PT) {
	core.borderSize = size
}

// SetMargins устанавливает границы документа для каждой из сторон в миллиметрах.
//
// По умолчанию: 3, 3, 3, 8.
func (core *Core) SetMargins(left, top, right, bottom unit.MM) {
	core.pagesConfig.SetMargins(left, top, right, bottom)
}

// ReadImage читает изображение по заданному пути и сохраняет его с заданным псевдонимом. Изображение должно быть формата PNG.
func (core *Core) ReadImage(path, alias string) error {
	if alias == "" {
		err := errors.New("псевдоним не может быть пустым")

		return err
	}

	imageBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	img, err := image.New(alias, imageBytes, core.compressor)
	if err != nil {
		return err
	}

	core.resources.SetImage(alias, img)

	return nil
}

// AddImage сохраняет бинарные данные изображения с заданным псевдонимом. Изображение должно быть формата PNG.
func (core *Core) AddImage(data []byte, alias string) error {
	if alias == "" {
		err := errors.New("псевдоним не может быть пустым")

		return err
	}

	img, err := image.New(alias, data, core.compressor)
	if err != nil {
		return err
	}

	core.resources.SetImage(alias, img)

	return nil
}

func (core *Core) DefaultFontSize() unit.PT {
	return core.fontSize
}

func (core *Core) DefaultBorderSize() unit.PT {
	return core.borderSize
}

func (core *Core) font(alias string) (*font.Font, error) {
	_font, err := core.resources.GetFont(alias)
	if err != nil {
		return nil, err
	}

	return _font, nil
}
