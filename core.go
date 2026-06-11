package pdf_craft

import (
	"compress/zlib"
	"errors"
	"fmt"
	"os"

	"github.com/eugene-static/pdf-craft/internal/buffer"
	"github.com/eugene-static/pdf-craft/internal/font"
	"github.com/eugene-static/pdf-craft/internal/image"
	"github.com/eugene-static/pdf-craft/pkg/unit"
)

const (
	FontRegular = "REG"
	FontBold    = "BOLD"
	FontItalic  = "ITALIC"

	Portrait  = "P"
	Landscape = "L"
)

// Core -- ядро конструктора. Позволяет настроить конструктор единожды и переиспользовать эти настройки при каждой новой генерацией.
// Так же содержит информацию, необходимую для различных узлов конструктора.
type Core struct {
	mainBuffer          *buffer.Buffer
	fonts               map[string]*font.Font
	images              map[string]*image.Image
	comp                compressor
	page                page
	fontAlias           string
	fontSize            unit.PT
	borderSize          unit.PT
	offsets             []int
	error               error
	compress            bool
	ignoreImageNotFound bool
}

// NewCore инициализирует новый экземпляр ядра конструктора с заранее заданной ориентацией страницы.
// "P" -- портретная ориентация (по умолчанию), "L" -- альбомная. Границы соответствуют формату А4.
func NewCore(orientation string) *Core {
	pg := page{
		buffer:       buffer.New(buffer.DefaultSize),
		width:        unit.PT(595.2).MM(),
		height:       unit.PT(841.89).MM(),
		marginLeft:   3,
		marginTop:    3,
		marginRight:  3,
		marginBottom: 8,
	}

	if orientation == Landscape {
		pg.width, pg.height = pg.height, pg.width
	}

	offsets := make([]int, 3, 100)

	compBuffer := buffer.New(buffer.DefaultSize)
	comp := compressor{
		buffer: compBuffer,
		writer: zlib.NewWriter(compBuffer),
	}

	return &Core{
		mainBuffer: buffer.New(buffer.DefaultSize),
		fonts:      make(map[string]*font.Font),
		images:     make(map[string]*image.Image),
		comp:       comp,
		page:       pg,
		fontSize:   unit.PT(6),
		borderSize: unit.PT(0.3),
		offsets:    offsets,
	}
}

// EnableCompression включает компрессию страниц документа. Значительно уменьшает объем файла, но увеличивает время на генерацию.
func (core *Core) EnableCompression() {
	core.compress = true
}

func (core *Core) DisableCompression() {
	core.compress = false
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

	f, err := font.New(path, alias)
	if err != nil {
		return err
	}

	core.fonts[alias] = f

	return nil
}

// SetFontRegular читает и устанавливает шрифт с псевдонимом "REG". При конфигурации ячейки указывать этот псевдоним не обязательно.
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
	core.page.marginLeft = left
	core.page.marginTop = top
	core.page.marginRight = right
	core.page.marginBottom = bottom
}

// PagesCount возвращает количество страниц на этапе построения документа равным очередности вызова функции.
func (core *Core) PagesCount() int {
	return core.page.count
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

	img, err := image.New(alias, imageBytes)
	if err != nil {
		return err
	}

	core.images[alias] = img

	return nil
}

// AddImage сохраняет бинарные данные изображения с заданным псевдонимом. Изображение должно быть формата PNG.
func (core *Core) AddImage(data []byte, alias string) error {
	if alias == "" {
		err := errors.New("псевдоним не может быть пустым")

		return err
	}

	img, err := image.New(alias, data)
	if err != nil {
		return err
	}

	core.images[alias] = img

	return nil
}

func (core *Core) defaultFontSize() unit.PT {
	return core.fontSize
}

func (core *Core) defaultBorderSize() unit.PT {
	return core.borderSize
}

func (core *Core) setError(err error) {
	core.error = err
}

func (core *Core) err() error {
	return core.error
}

func (core *Core) bytes() []byte {
	return core.mainBuffer.Bytes()
}

func (core *Core) font(alias string) *font.Font {
	if len(core.fonts) == 0 {
		err := errors.New("нет установленных шрифтов")

		core.setError(err)

		return nil
	}

	fnt, ok := core.fonts[alias]
	if !ok {
		err := fmt.Errorf("не найден шрифт с таким именем: %s", alias)

		core.setError(err)

		return nil
	}

	return fnt
}

func (core *Core) image(alias string) *image.Image {
	img, ok := core.images[alias]
	if !ok && !core.ignoreImageNotFound {
		err := fmt.Errorf("не найдено изображение с таким именем: %s", alias)

		core.setError(err)

		return nil
	}

	return img
}

func (core *Core) newObject() int {
	objNum := len(core.offsets)

	xLen := core.mainBuffer.Len()

	core.offsets = append(core.offsets, xLen)

	return objNum
}

func (core *Core) setObject(objNum int) {
	xLen := core.mainBuffer.Len()

	core.offsets[objNum] = xLen
}

func (core *Core) newPage() {
	if core.err() != nil {
		return
	}

	if core.page.headerBuffer != nil && core.page.headerBuffer.Len() > 0 {
		_, err := core.page.buffer.Write(core.page.headerBuffer.Bytes())
		if err != nil {
			core.setError(err)

			return
		}
	}

	core.page.count++
}

func (core *Core) renderPage() {
	if core.err() != nil {
		return
	}

	if core.page.watermarkBuffer != nil && core.page.watermarkBuffer.Len() > 0 {
		_, err := core.page.buffer.Write(core.page.watermarkBuffer.Bytes())
		if err != nil {
			core.setError(err)

			return
		}
	}

	if core.page.buffer.Len() > 0 {
		core.writePage()

		core.page.buffer.Reset()
	}
}

func (core *Core) reset() {
	core.offsets = core.offsets[:3]
	core.error = nil
	core.mainBuffer.Reset()
}

type page struct {
	buffer          *buffer.Buffer
	headerBuffer    *buffer.Buffer
	watermarkBuffer *buffer.Buffer
	objects         []int
	count           int
	width           unit.MM
	height          unit.MM
	marginLeft      unit.MM
	marginRight     unit.MM
	marginTop       unit.MM
	marginBottom    unit.MM
}

func (p *page) x0y0() (unit.MM, unit.MM) {
	return p.x0(), p.y0()
}

func (p *page) x0() unit.MM {
	return p.marginLeft
}

func (p *page) y0() unit.MM {
	return p.marginTop - p.height
}

func (p *page) yB() unit.MM {
	return -p.marginBottom
}

func (p *page) isBelowBottomBorder(y unit.MM) bool {
	return y+p.marginBottom > 0
}

func (p *page) newHeader() *buffer.Buffer {
	if p.headerBuffer != nil {
		p.headerBuffer.Reset()

		return p.headerBuffer
	}

	buf := buffer.New(buffer.DefaultSize)

	p.headerBuffer = buf

	return buf
}

func (p *page) releaseHeader() {
	if p.headerBuffer != nil {
		p.headerBuffer.Reset()
	}
}

func (p *page) newWatermark() *buffer.Buffer {
	if p.watermarkBuffer != nil {
		p.watermarkBuffer.Reset()

		return p.watermarkBuffer
	}

	buf := buffer.New(buffer.DefaultSize)

	p.watermarkBuffer = buf

	return buf
}

func (p *page) releaseWatermark() {
	if p.watermarkBuffer != nil {
		p.watermarkBuffer.Reset()
	}
}

func (p *page) reset() {
	p.releaseWatermark()
	p.releaseHeader()
	p.count = 0
	p.objects = p.objects[:0]
}

type compressor struct {
	buffer *buffer.Buffer
	writer *zlib.Writer
}

func (comp compressor) compress(buf []byte) ([]byte, error) {
	comp.buffer.Reset()
	comp.writer.Reset(comp.buffer)

	_, err := comp.writer.Write(buf)
	if err != nil {
		return nil, err
	}

	err = comp.writer.Close()
	if err != nil {
		return nil, err
	}

	result := make([]byte, comp.buffer.Len())
	copy(result, comp.buffer.Bytes())

	return result, nil
}

type ImageNotFoundError struct {
	alias string
}

func (err *ImageNotFoundError) Error() string {
	return fmt.Sprintf("не найдено изображение с таким именем: %s", err.alias)
}
