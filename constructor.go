package pdfego

import (
	"errors"
	"os"
	"time"

	"github.com/eugene-static/pdfego/internal/components/object/pages"
	"github.com/eugene-static/pdfego/internal/components/object/resources"
	"github.com/eugene-static/pdfego/internal/components/object/resources/font"
	"github.com/eugene-static/pdfego/internal/components/object/resources/image"
	"github.com/eugene-static/pdfego/internal/components/parameter"
	"github.com/eugene-static/pdfego/internal/components/stream"
	"github.com/eugene-static/pdfego/internal/components/stream/bytes"
	"github.com/eugene-static/pdfego/internal/components/stream/primitives"
	"github.com/eugene-static/pdfego/internal/errs"
	"github.com/eugene-static/pdfego/internal/manager"
	"github.com/eugene-static/pdfego/unit"
)

// Columns возвращает слайс значений в миллиметрах. Функция удобна для задания ширин колонок таблицы.
func Columns(columns ...unit.MM) []unit.MM {
	return columns
}

// Constructor -- главный узел дерева построения макета документа.
type Constructor struct {
	core      *Core
	mgr       *manager.Manager
	paginator *Paginator
	block     Block
	origin    unit.Point
	offset    unit.Point
	cursor    unit.Point
}

// NewConstructor создает новый экземпляр конструктора с заданным ядром. Задает первую страницу макета.
func NewConstructor(core *Core) *Constructor {
	mgr := manager.New(core.pagesConfig)

	mgr.EnableCompression(core.enableCompression)
	mgr.StartDocument()
	mgr.NewPage()

	paginator := &Paginator{
		counter: &counter{
			config: make(parameter.Numbers, 2, 4),
		},
		pages: mgr.Pages(),
	}

	block := Block{
		core:    core,
		ctx:     mgr.Context(),
		res:     mgr.Resources(),
		counter: paginator.counter,
	}

	viewBox := mgr.ViewBox()
	origin := unit.NewPoint(viewBox.X0, viewBox.Y0)

	return &Constructor{
		core:      core,
		mgr:       mgr,
		paginator: paginator,
		block:     block,
		origin:    origin,
		offset:    origin,
		cursor:    origin,
	}
}

// Build использует функцию BlockApplier для построения одного блока.
//
// Пример:
//
//	func title(block *Block) {
//		block.Slot().Table(1, Columns(20)).
//			Row().Cell("Документ", CellOptions{Font: "BOLD", FontSize: 14, Align: "CT"}
//	}
//
//	func Fill() {
//		...
//		constructor.Build(title)
//		...
func (c *Constructor) Build(applier BlockApplier, options ...NodeOptions) *Constructor {
	c.newBlock(options, orderMode)

	applier(&c.block)

	return c
}

// NewPage создает новую страницу.
func (c *Constructor) NewPage() {
	c.render()
	c.renderPage()

	c.mgr.NewPage()
	c.mgr.ResetBuffers()

	c.cursor = c.offset
}

// Bytes завершает документ и отдает бинарные данные готового документа и ошибку, если такова была во время создания и рендеринга макета.
func (c *Constructor) Bytes() ([]byte, error) {
	c.render()
	c.renderPage()
	c.renderPagesCount()

	c.mgr.FinishDocument(c.core.resources)

	err := c.mgr.Context().Error()
	if err != nil {
		return nil, err
	}

	return c.mgr.Bytes(), nil
}

// ReadImage читает изображение по заданному пути и сохраняет его с заданным псевдонимом.
// Изображение будет использовано только для текущей итерации.
// Изображение должно быть формата PNG.
func (c *Constructor) ReadImage(path, alias string) error {
	if alias == "" {
		err := errors.New("псевдоним не может быть пустым")

		return err
	}

	imageBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	img, err := image.New(alias, imageBytes, c.mgr.Compressor())
	if err != nil {
		return err
	}

	c.mgr.Resources().SetImage(alias, img)

	return nil
}

// AddImage сохраняет бинарные данные изображения с заданным псевдонимом.
// Изображение будет использовано только для текущей итерации.
// Изображение должно быть формата PNG.
func (c *Constructor) AddImage(data []byte, alias string) error {
	if alias == "" {
		err := errors.New("псевдоним не может быть пустым")

		return err
	}

	img, err := image.New(alias, data, c.mgr.Compressor())
	if err != nil {
		return err
	}

	c.mgr.Resources().SetImage(alias, img)

	return nil
}

// SetProducer устанавливает автора документа. По-умолчанию: "pdfego".
func (c *Constructor) SetProducer(producer string) {
	c.mgr.Info().SetProducer(producer)
}

// SetCreationDate устанавливает время создания документа. По-умолчанию: time.Now().
func (c *Constructor) SetCreationDate(date time.Time) {
	c.mgr.Info().SetCreationDate(date)
}

// Возвращает экземпляр Block. Все блоки располагаются вертикально друг за другом.
// При каждом новом вызове метода, предыдущий экземпляр рендерится и больше не может быть изменен.
func (c *Constructor) Block(options ...NodeOptions) *Block {
	c.newBlock(options, orderMode)

	return &c.block
}

// Возвращает экземпляр Header. Этот блок будет повторяться в начале каждой новой страницы с момента инициализации
// и до вызова метода ReleaseHeader(). Так же он отрисуется в момент инициализации.
func (c *Constructor) Header(options ...NodeOptions) *Header {
	c.newBlock(options, orderMode)

	return &Header{
		block:  &c.block,
		offset: &c.offset,
		mgr:    c.mgr,
	}
}

// ApplyHeader -- инлайн-обретка над последовательностью Constructor.Header() и Header.Apply(BlockApplier)
func (c *Constructor) ApplyHeader(applier BlockApplier, options ...NodeOptions) *Constructor {
	c.newBlock(options, orderMode)

	applier(&c.block)

	wb := c.mgr.ViewBox()
	cursor := unit.NewPoint(wb.X0, wb.Y0)

	c.block.render(c.mgr.HeaderStream(), cursor)
	c.offset.AddY(c.block.height())
	c.mgr.WriteHeader()

	return c
}

// ReleaseHeader останавливает повторение блока Header на каждой странице.
func (c *Constructor) ReleaseHeader() *Constructor {
	c.mgr.HeaderRelease()

	c.offset.SetY(c.origin.Y())

	return c
}

// Repeater инициализирует повторяющийся блок.
// Deprecated. Используйте NewIterator
func (c *Constructor) Repeater(sectioner Sectioner, options ...NodeOptions) *Repeater {
	c.newBlock(options, 0)

	return &Repeater{
		block:      &c.block,
		sectioner:  sectioner,
		renderFunc: c.render,
	}
}

// Iterate использует Iterator для рендера повторяющегося блока. Количество повторений соответствует размеру коллекции итератора.
func (c *Constructor) Iterate(iterator iterator, options ...NodeOptions) *Constructor {
	iterator.forEach(&c.block, func(index int) {
		if index == 0 {
			c.newBlock(options, iteratorMode)

			return
		}

		c.render()
	})

	return c
}

// Возвращает экземпляр Watermark. Этот блок рендерится на каждой странице с момента инициализации.
// В отличие от Header не влияет на расположение других узлов конструктора.
func (c *Constructor) Watermark(options ...WatermarkOptions) *Watermark {
	opts := getOptions(options)
	alignH, alignV := parseAlignment(opts.Align)

	return &Watermark{
		block: &Block{
			core:      c.core,
			ctx:       c.mgr.Context(),
			res:       c.mgr.Resources(),
			slotIndex: zeroIndex,
			mode:      orderMode,
		},
		mgr:    c.mgr,
		alignH: alignH,
		alignV: alignV,
	}
}

// Возвращает экземпляр Paginator. Этот блок рендерится на каждой странице с момента инициализации.
// Располагается в одном из колонтитулов документа -- в области между отступом и границей документа.
// Пишет номер текущей страницы в первую ячейку с ID = 0.
func (c *Constructor) Paginator(options ...PaginatorOptions) *Paginator {
	opts := getOptions(options)
	alignH, alignV := parseAlignment(opts.Align)
	placement := parsePlacement(opts.Placement)

	block := &Block{
		core:      c.core,
		ctx:       c.mgr.Context(),
		res:       c.mgr.Resources(),
		slotIndex: zeroIndex,
		mode:      iteratorMode,
	}

	c.paginator.block = block
	c.paginator.offset = opts.Offset
	c.paginator.skip = opts.Skip
	c.paginator.alignH = alignH
	c.paginator.alignV = alignV
	c.paginator.placement = placement
	c.paginator.enabled = true

	return c.paginator
}

// Paginate создает блок для отображения номера страницы.
func (c *Constructor) Paginate(offset int) {
	c.
		Paginator(PaginatorOptions{Offset: offset, Align: "CT"}).
		Apply(DefaultPaginator)
}

// Новый блок не создается. Вся информация, записанная в блок рендерится и записывается в буфер страницы.
// После блок сбрасывается и к нему применяются новые опции. Поэтому нельзя изменять "старый блок" после вызова "нового".
func (c *Constructor) newBlock(options []NodeOptions, mode mode) {
	opts := getOptions(options)

	c.render()

	c.block.slots = c.block.slots[:0]
	c.block.slotIndex = zeroIndex
	c.block.mode = mode
	c.block.indentH = opts.IndentH
	c.block.indentV = opts.IndentV
	c.block.ledge = opts.Ledge
	c.block.spacing = opts.Spacing
	c.block.border = parseBorder(opts.Border)
	c.block.borderSize = opts.BorderSize
}

func (c *Constructor) render() {
	dy := c.block.height().Add(c.block.indentV)

	if c.mgr.IsBelowBottomBorder(c.cursor.Y().Add(dy)) {
		c.renderPage()
		c.mgr.ResetBuffers()
		c.mgr.NewPage()

		c.cursor = c.offset
	}

	c.block.render(c.mgr.PageStream(), c.cursor)
	c.block.reset()

	c.cursor.AddY(dy)
}

func (c *Constructor) renderPage() {
	if c.paginator.enabled {
		c.paginator.render()
	}

	c.mgr.WritePage()
}

func (c *Constructor) renderPagesCount() {
	if c.paginator.counter.cell != nil {
		c.paginator.renderCounter(c.mgr.Context())

		pc, err := c.mgr.Pages().PagesCount(c.paginator.counter.config)
		if err != nil {
			c.mgr.Context().SetError(err)

			return
		}

		c.mgr.Resources().SetObject(pages.AliasPagesCount, pc)
	}
}

// Block -- экземпляр блока. Все блоки располагаются друг за другом вертикально, независимо от того, кем был создан экземпляр.
// Блок в составе конструктора неделим: если его высота больше оставшегося места на странице, он будет отображен на следующей странице.
type Block struct {
	core       *Core
	ctx        *manager.Context
	res        *resources.Resources
	counter    *counter
	slots      []Slot
	slotIndex  uint8
	mode       mode
	border     uint8
	borderSize unit.PT
	indentH    unit.MM
	indentV    unit.MM
	ledge      unit.MM
	spacing    unit.MM
}

type BlockApplier func(block *Block)

// Apply вызывает функцию func(block *Block).
func (b *Block) Apply(applier BlockApplier) {
	applier(b)
}

// Возвращает экземпляр Slot.
func (b *Block) Slot(options ...NodeOptions) *Slot {
	b.slotIndex++

	if b.slotIndex < uint8(len(b.slots)) && b.slots[b.slotIndex].mode.is(iteratorMode) {
		b.slots[b.slotIndex].reset()

		return &b.slots[b.slotIndex]
	}

	opts := getOptions(options)

	s := Slot{
		core:       b.core,
		ctx:        b.ctx,
		res:        b.res,
		counter:    b.counter,
		blockIndex: zeroIndex,
		mode:       b.mode,
		indentH:    opts.IndentH,
		indentV:    opts.IndentV,
		spacing:    opts.Spacing,
		ledge:      opts.Ledge,
		borderSize: opts.BorderSize,
		border:     parseBorder(opts.Border),
	}

	b.slots = append(b.slots, s)

	return &b.slots[b.slotIndex]
}

func (b *Block) render(dst *stream.Stream, cursor unit.Point) {
	cursor.AddX(b.indentH)
	cursor.AddY(b.indentV)

	if b.border > 0 {
		width := b.width()
		height := b.height()
		borderSize := coalesce(b.borderSize, b.core.DefaultBorderSize())

		renderBorder(dst, cursor, width, height, b.border, borderSize)
	}

	for i := range b.slots {
		b.slots[i].render(dst, cursor)

		cursor.AddX(b.slots[i].width() + b.slots[i].indentH + b.spacing)
	}
}

func (b *Block) modify(section []string) {
	for i := range b.slots {
		b.slots[i].modify(section)
	}
}

func (b *Block) height() (h unit.MM) {
	for i := range b.slots {
		sh := b.slots[i].height() + b.slots[i].indentV
		if h < sh {
			h = sh
		}
	}

	return h + b.ledge
}

func (b *Block) width() (w unit.MM) {
	for i := range b.slots {
		w += b.slots[i].width() + b.slots[i].indentH + b.spacing
	}

	w -= b.spacing

	return w
}

func (b *Block) reset() {
	b.slotIndex = zeroIndex
}

// Header -- блок, повторяющийся в начале каждой страницы. Он является частью макета документа,
// поэтому не может быть изменен после создания нового блока.
type Header struct {
	block  *Block
	mgr    *manager.Manager
	offset *unit.Point
}

// Apply вызывает переданную функцию и сохраняет результат отрисовки в собственном буфере.
func (h *Header) Apply(applier BlockApplier) {
	applier(h.block)

	wb := h.mgr.ViewBox()
	cursor := unit.NewPoint(wb.X0, wb.Y0)

	h.block.render(h.mgr.HeaderStream(), cursor)
	h.offset.AddY(h.block.height())
	h.mgr.WriteHeader()
}

// Repeater -- повторяющийся блок. Количество повторений задается через интерфейс Sectioner.
// Он является частью макета документа, поэтому не может быть изменен после создания нового блока.
type Repeater struct {
	block      *Block
	sectioner  Sectioner
	renderFunc func()
}

type RepeaterApplier func(block *Block, section []string)

// Sectioner -- интерфейс взаимодействия с Repeater.
// Метод Section должен возвращать слайс строк, где индекс строки будет соответствовать индексу ячейки, в которую будет записываться значение этой строки.
// Метод Count должен возвращать количество таких секций и будет соответствовать количеству повторений.
//
// Пример:
//
//	type Employee struct {
//		ID   int
//		Name string
//	}
//
//	type Employees struct {
//		Location  string
//		Employees []Employee
//	}
//
//	func (emp *Employees) Section(index int) []string {
//		if index >= len(emp.Employees) {
//			return nil
//		}
//
//		section := make([]string, 2)
//
//		section[0] = strconv.Itoa(emp.Employees[index].ID)
//		section[1] = emp.Employees[index].Name
//
//		return section
//	}
//
//	func (emp *Employees) Count() int {
//		return len(emp.Employees)
//	}
//
//	func employeesTable(block *Block, section []string) {
//		table := block.Slot().Table(2, Columns(10, 20))
//
//		table.Row().
//			Cell(section[0], CellOptions{ID: 0, Border: "o"}).
//			Cell(section[1], CellOptions{ID: 1, Border: "o"})
//	}
//
//	func fill(core *Core) {
//		constructor := NewConstructor(core)
//
//		employees := &Employees{
//			Location: "Moscow",
//			Employees: []Employee{
//				{
//					ID: 123,
//					Name: "Ярослав Дронов",
//				},
//				{
//					ID: 129,
//					Name: "Сергей Пенкин",
//				},
//			},
//		}
//
//		constructor.Repeater(employees).Repeat(employeesTable)
//	}
type Sectioner interface {
	Section(int) []string
	Count() int
}

// Repeat повторит переданный блок sectioner.Count() раз. Макет блока не перестраивается при каждом повторении,
// а модифицируется в зависимости от переданного текста.
func (r *Repeater) Repeat(applier RepeaterApplier) {
	sectionsCount := r.sectioner.Count()
	if sectionsCount <= 0 {
		//r.block.reset()

		return
	}

	sectionIndex := 0
	section := r.sectioner.Section(sectionIndex)

	applier(r.block, section)

	for i := sectionIndex + 1; i < sectionsCount; i++ {
		r.renderFunc()

		section = r.sectioner.Section(i)

		r.block.modify(section)
	}
}

// Iterator хранит сведения о повторяющемся блоке. Используется для метода Constructor.Iterate
type Iterator[T any] struct {
	collection []T
	applier    IteratorApplier[T]
}

type IteratorApplier[T any] func(block *Block, item T, index int)

type iterator interface {
	forEach(block *Block, fn func(int))
}

// NewIterator инициализирует новый Iterator. Для инициализации необходима коллекция данных (напр. слайс структур) и функция-шаблон для одного элемента коллекции.
func NewIterator[T any](collection []T, applier IteratorApplier[T]) *Iterator[T] {
	return &Iterator[T]{
		collection: collection,
		applier:    applier,
	}
}

func (it *Iterator[T]) forEach(block *Block, fn func(index int)) {
	for i := range it.collection {
		fn(i)
		it.applier(block, it.collection[i], i)
	}
}

// Watermark -- блок, повторяющейся на каждой странице. Водяной знак не является частью основного макета документа и будет располагаться поверх него.
// Позиционирование блока происходит в пределах границ страницы.
type Watermark struct {
	block  *Block
	mgr    *manager.Manager
	alignH uint8
	alignV uint8
}

// Apply вызывает переданную функцию и сохраняет результат отрисовки в собственном буфере.
func (w *Watermark) Apply(applier BlockApplier) {
	applier(w.block)

	wb := w.mgr.ViewBox()
	dx := w.dx(wb)
	dy := w.dy(wb)
	cursor := unit.NewPoint(wb.X0+dx, wb.Y0+dy)

	w.block.render(w.mgr.WatermarkStream(), cursor)

	w.mgr.WriteWatermark()
}

func (w *Watermark) dx(wb pages.ViewBox) (dx unit.MM) {
	switch w.alignH {
	case alignR:
		dx = wb.Width - w.block.width()
	case alignM:
		dx = (wb.Width - w.block.width()) / 2
	default:
		dx = 0
	}

	return dx
}

func (w *Watermark) dy(wb pages.ViewBox) (dy unit.MM) {
	switch w.alignV {
	case alignB:
		dy = wb.Height - w.block.height()
	case alignM:
		dy = (wb.Height - w.block.height()) / 2
	default:
		dy = 0
	}

	return dy
}

// Paginator -- блок, повторяющийся на каждой странице между ее нижней границей и ее нижним краем.
// Пагинатор не является частью основного макета документа. Макет блока не перестраивается при каждом повторении,
// а модифицируется в зависимости от текста номера страницы.
type Paginator struct {
	block     *Block
	pages     *pages.Pages
	counter   *counter
	applier   PaginatorApplier
	offset    int
	skip      uint
	alignH    uint8
	alignV    uint8
	placement uint8
	enabled   bool
}

// Создает экземпляр Slot аналогично такому же методу у Block.
func (p *Paginator) Slot(options ...NodeOptions) *Slot {
	return p.block.Slot(options...)
}

// Apply вызывает переданную функцию.
func (p *Paginator) Apply(applier PaginatorApplier) *Paginator {
	p.applier = applier

	return p
}

// Disable отключает пагинатор.
func (p *Paginator) Disable() {
	p.enabled = false
}

type PaginatorApplier func(block *Block, number string)

func DefaultPaginator(block *Block, number string) {
	block.Slot().Table(1, Columns(5)).Row().Cell(number, CellOptions{TextColor: ColorGray50})
}

func (p *Paginator) render() {
	pagesCount := p.pages.Count()
	if pagesCount <= int(p.skip) {
		return
	}

	number := bytes.FormatInt(pagesCount + p.offset)

	p.applier(p.block, number)

	cfg := p.pages.Config()
	x0 := cfg.MarginLeft
	y0 := cfg.UpperRightY.Neg()

	if p.placement == placementBottom {
		y0 = cfg.MarginBottom.Neg()
	}

	dx := p.dx(cfg)
	dy := p.dy(cfg)

	cursor := unit.NewPoint(x0+dx, y0+dy)

	p.block.render(p.pages.PageStream(), cursor)
	p.block.reset()
}

func (p *Paginator) renderCounter(ctx *manager.Context) {
	count := bytes.FormatInt(p.pages.Count())
	cursor := unit.NewPoint(0, 0)

	p.counter.cell.setTextLines(ctx, count)
	p.counter.cell.renderText(p.pages.CountStream(), cursor, ColorDefault)
}

func (p *Paginator) dx(cfg pages.Config) (dx unit.MM) {
	width := cfg.UpperRightX - cfg.MarginRight

	switch p.alignH {
	case alignR:
		dx = width - p.block.width()
	case alignM:
		dx = (width - p.block.width()) / 2
	default:
		dx = 0
	}

	return dx
}

func (p *Paginator) dy(cfg pages.Config) (dy unit.MM) {
	area := cfg.MarginTop
	if p.placement == placementBottom {
		area = cfg.MarginBottom
	}

	switch p.alignV {
	case alignB:
		dy = area - p.block.height()
	case alignM:
		dy = (area - p.block.height()) / 2
	default:
		dy = 0
	}

	return dy
}

type counter struct {
	cell   *cell
	config parameter.Numbers
}

func (pc *counter) register(cell *cell) {
	c := *cell

	pc.cell = &c
}

// Slot представляет собой слот. Слоты располагаются друг за другом горизонтально. Может содержать либо дочерние блоки Block, либо таблицу Table.
// Высота родительского блока будет являться высотой самого высокого слота.
type Slot struct {
	core       *Core
	ctx        *manager.Context
	res        *resources.Resources
	counter    *counter
	blocks     []Block
	table      *Table
	mode       mode
	blockIndex uint8
	border     uint8
	borderSize unit.PT
	indentH    unit.MM
	indentV    unit.MM
	ledge      unit.MM
	spacing    unit.MM
}

type SlotApplier func(slot *Slot)

// Создает дочерний экземпляр блока Block. Если к слоту была добавлена таблица Table, то конструктор вернет ошибку, а блоки не будут отрисованы.
// Ширина слота будет являться шириной самого широкого блока. Высота слота будет суммой высот всех блоков.
func (s *Slot) Block(options ...NodeOptions) *Block {
	if s.table != nil {
		err := errors.New("слот уже содержит таблицу")

		s.ctx.SetError(err)
	}

	s.blockIndex++

	if s.blockIndex < uint8(len(s.blocks)) && s.blocks[s.blockIndex].mode.is(iteratorMode) {
		s.blocks[s.blockIndex].reset()

		return &s.blocks[s.blockIndex]
	}

	opts := getOptions(options)

	b := Block{
		core:       s.core,
		ctx:        s.ctx,
		res:        s.res,
		counter:    s.counter,
		slotIndex:  zeroIndex,
		mode:       s.mode,
		indentH:    opts.IndentH,
		indentV:    opts.IndentV,
		ledge:      opts.Ledge,
		spacing:    opts.Spacing,
		border:     parseBorder(opts.Border),
		borderSize: opts.BorderSize,
	}

	s.blocks = append(s.blocks, b)

	return &s.blocks[s.blockIndex]
}

func (s *Slot) reset() {
	s.blockIndex = zeroIndex
}

// Создает дочерний экземпляр таблицы Table. Таблица у слота может быть только одна. При повторном вызове метода, таблица перезапишется на новую.
// Для инициализации таблицы необходимо указать количество строк и слайс ширин всех столбцов.
func (s *Slot) Table(rowsNum uint8, columns []unit.MM, options ...TableOptions) *Table {
	if len(s.blocks) > 0 {
		err := errors.New("слот уже содержит блоки")

		s.ctx.SetError(err)
	}

	if s.table != nil && s.table.mode.is(iteratorMode) {
		s.table.reset()

		return s.table
	}

	opts := getOptions(options)

	columnsNum := uint8(len(columns))

	s.table = &Table{
		core:        s.core,
		ctx:         s.ctx,
		res:         s.res,
		counter:     s.counter,
		mode:        s.mode,
		columns:     columns,
		rows:        make([]Row, rowsNum),
		rowspans:    make([]uint8, columnsNum),
		cellsPool:   make([]cell, rowsNum*columnsNum),
		textColor:   opts.TextColor,
		borderColor: opts.BorderColor,
		border:      parseBorder(opts.Border),
		borderSize:  opts.BorderSize,
		indentH:     opts.IndentH,
		indentV:     opts.IndentV,
		spacingH:    opts.SpacingH,
		spacingV:    opts.SpacingV,
	}

	return s.table
}

// Apply вызывает переданную функцию.
func (s *Slot) Apply(applier SlotApplier) *Slot {
	applier(s)

	return s
}

func (s *Slot) render(dst *stream.Stream, cursor unit.Point) {
	cursor.AddX(s.indentH)
	cursor.AddY(s.indentV)

	if s.border > 0 {
		width := s.width()
		height := s.height()
		borderSize := coalesce(s.borderSize, s.core.DefaultBorderSize())

		renderBorder(dst, cursor, width, height, s.border, borderSize)
	}

	if s.table != nil {
		s.table.render(dst, cursor)

		return
	}

	for i := range s.blocks {
		s.blocks[i].render(dst, cursor)

		cursor.AddY(s.blocks[i].height() + s.blocks[i].indentV + s.spacing)
	}
}

func (s *Slot) modify(section []string) {
	if s.table != nil {
		s.table.modify(section)
	}

	for i := range s.blocks {
		s.blocks[i].modify(section)
	}
}

func (s *Slot) height() (h unit.MM) {
	h = s.ledge

	if s.table != nil {
		h += s.table.height() + s.table.indentV

		return h
	}

	for i := range s.blocks {
		h += s.blocks[i].height() + s.blocks[i].indentV + s.spacing
	}

	h -= s.spacing

	return h
}

func (s *Slot) width() (w unit.MM) {
	if s.table != nil {
		w = s.table.width() + s.table.indentH

		return w
	}

	for i := range s.blocks {
		bw := s.blocks[i].width() + s.blocks[i].indentH
		if w < bw {
			w = bw
		}
	}

	return w
}

// Table представляет собой таблицу.
type Table struct {
	core        *Core
	ctx         *manager.Context
	res         *resources.Resources
	counter     *counter
	mode        mode
	columns     []unit.MM
	rows        []Row
	cellsPool   []cell
	rowspans    []uint8
	rowIndex    uint8
	textColor   Color
	borderColor Color
	border      uint8
	borderSize  unit.PT
	indentH     unit.MM
	indentV     unit.MM
	spacingH    unit.MM
	spacingV    unit.MM
}

type TableApplier func(*Table)

// Apply вызывает переданную функцию.
func (t *Table) Apply(applier TableApplier) {
	applier(t)
}

// Row представляет собой строку таблицы Table. При попытке вызова метода большее число раз, чем количество строк в таблице,
// будет перезаписываться последняя строка.
//
// Возможно задать минимальную высоту строки.
func (t *Table) Row(options ...RowOptions) *Row {
	opts := getOptions(options)

	row := t.row()

	defer t.updateIndexes()

	if row.mode.is(iteratorMode) {
		t.resetRow(row, opts)

		return row
	}

	t.initRow(row, opts)

	return row
}

func (t *Table) row() *Row {
	row := &t.rows[t.rowIndex]

	return row
}

func (t *Table) initRow(row *Row, options RowOptions) {
	columnsLen := len(t.columns)
	start := int(t.rowIndex) * columnsLen
	end := start + columnsLen

	row.core = t.core
	row.ctx = t.ctx
	row.res = t.res
	row.mode = t.mode
	row.counter = t.counter
	row.cells = t.cellsPool[start:end]
	row.columns = t.columns
	row.height = options.MinHeight
	row.columnsLen = uint8(columnsLen)
	row.rowspans = t.rowspans
	row.spacing = t.spacingH
	row.columnIndex = 0
}

func (t *Table) resetRow(row *Row, options RowOptions) {
	row.height = options.MinHeight
	row.columnIndex = 0
}

func (t *Table) render(dst *stream.Stream, cursor unit.Point) {
	if t.ctx.IsError() {
		return
	}

	cursor.AddX(t.indentH)
	cursor.AddY(t.indentV)

	if !t.textColor.isDefault() {
		t.textColor.WriteToStream(dst, primitives.FillColorRGB)
		defer ColorDefault.WriteToStream(dst, primitives.FillColorRGB)
	}

	if t.border > 0 {
		if !t.borderColor.isDefault() {
			t.borderColor.WriteToStream(dst, primitives.StrokeColorRGB)
			defer ColorDefault.WriteToStream(dst, primitives.StrokeColorRGB)
		}

		width := t.width()
		height := t.height()
		borderSize := coalesce(t.borderSize, t.core.DefaultBorderSize())

		renderBorder(dst, cursor, width, height, t.border, borderSize)
	}

	cc := cursor

	for i := range t.cellsPool {
		ri := i / len(t.columns)
		ci := i % len(t.columns)

		if !t.cellsPool[i].mode.is(skipMode) {
			t.setCellHeight(ri, i)

			t.cellsPool[i].render(dst, cc, t.textColor, t.borderColor)

			if t.cellsPool[i].mode.is(counterMode) {
				t.counter.config = append(t.counter.config,
					parameter.Number(t.cellsPool[i].width.PT()),
					parameter.Number(t.cellsPool[i].height.PT()),
				)
			}
		}

		cc.AddX(t.columns[ci] + t.spacingH)

		if ci == len(t.columns)-1 {
			cc.SetX(cursor.X())
			cc.AddY(t.rows[ri].height + t.spacingV)
		}
	}

	t.rowIndex = 0
}

func (t *Table) modify(section []string) {
	for i := range t.cellsPool {
		ri := i / len(t.columns)
		ci := i % len(t.columns)

		if ci == 0 {
			t.rows[ri].height = 0
		}

		id := t.cellsPool[i].id
		if int(id) > len(section)-1 {
			continue
		}

		t.rows[ri].modifyCell(&t.cellsPool[i], section[id])
	}
}

func (t *Table) height() (h unit.MM) {
	for i := range t.rows {
		h += t.rows[i].height + t.spacingV
	}

	h -= t.spacingV

	return h
}

func (t *Table) width() (w unit.MM) {
	for i := range t.columns {
		w += t.columns[i] + t.spacingH
	}

	w -= t.spacingH

	return w
}

// Выставляем высоту ячейки по высоте строк, которые она занимает.
func (t *Table) setCellHeight(ri, cpi int) {
	height := unit.MM(0)

	for i := ri; i < len(t.rows) && i < ri+int(t.cellsPool[cpi].rowspan); i++ {
		height += t.rows[i].height
	}

	t.cellsPool[cpi].height = height
}

func (t *Table) updateIndexes() {
	if t.rowIndex+1 < uint8(len(t.rows)) {
		t.rowIndex++
	}

	for i := range t.rowspans {
		if t.rowspans[i] > 0 {
			t.rowspans[i]--
		}
	}
}

func (t *Table) reset() {
	for i := range t.rowspans {
		t.rowspans[i] = 0
	}

	t.rowIndex = 0
}

// Row представляет собой экземпляр строки таблицы Table. На расстояние между строками влияет значение TableOptions.SpacingV.
type Row struct {
	core        *Core
	ctx         *manager.Context
	res         *resources.Resources
	counter     *counter
	cells       []cell
	columns     []unit.MM
	rowspans    []uint8
	height      unit.MM
	spacing     unit.MM
	columnIndex uint8
	columnsLen  uint8
	mode        mode
}

type RowApplier func(*Row)

func (r *Row) Apply(applier RowApplier) {
	applier(r)
}

// Cell -- ячейка таблицы с заданным текстом. Использует шрифт, цвет по-умолчанию. Без рамки, позиционирование CM.
func (r *Row) Cell(text string, options ...CellOptions) *Row {
	opts := getOptions(options)

	r.textCell(text, opts)

	return r
}

// Label -- ячейка таблицы с заданным текстом. Без рамки, позиционирование LB.
func (r *Row) Label(text string, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, "LB")

	r.textCell(text, opts)

	return r
}

// LabelSpan -- Label с заданным Colspan.
func (r *Row) LabelSpan(text string, colspan uint8, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, "LB")
	opts.Colspan = colspan

	r.textCell(text, opts)

	return r
}

// LabelHead -- Label, где в качестве шрифта используется заданный шрифт с псевдонимом "BOLD".
func (r *Row) LabelHead(text string, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, "LB")
	opts.Font = coalesce(opts.Font, r.core.fontBoldAlias)

	r.textCell(text, opts)

	return r
}

// Blank -- ячейка таблицы, бланк, с заданным текстом и позиционированием. Тонкая граница снизу и позиционирование по-центру-снизу по-умолчанию.
func (r *Row) Blank(text, align string, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, align, "CB")
	opts.Border = coalesce(opts.Border, "b")

	r.textCell(text, opts)

	return r
}

// BlankSpan -- Blank с заданным Colspan.
func (r *Row) BlankSpan(text, align string, colspan uint8, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, align, "CB")
	opts.Border = coalesce(opts.Border, "b")
	opts.Colspan = colspan

	r.textCell(text, opts)

	return r
}

// BlankEmpty -- пустой Blank, без текста.
func (r *Row) BlankEmpty(options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Border = coalesce(opts.Border, "b")

	r.textCell("", opts)

	return r
}

// Form -- Blank с переносом текста.
func (r *Row) Form(text, align string, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, align, "CB")
	opts.Border = coalesce(opts.Border, "b")
	opts.Wrap = true

	r.textCell(text, opts)

	return r
}

// FormSpan -- BlankSpan с переносом текста.
func (r *Row) FormSpan(text, align string, colspan uint8, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, align, "CB")
	opts.Border = coalesce(opts.Border, "b")
	opts.Colspan = colspan
	opts.Wrap = true

	r.textCell(text, opts)

	return r
}

// Paragraph -- ячейка таблицы с заданным текстом. Позиционирование "CB"
func (r *Row) Paragraph(text string, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, "CB")

	r.textCell(text, opts)

	return r
}

// Underscore -- ячейка таблицы с заданным текстом. Размер шрифта меньше заданного по-умолчанию на 1 пункт.
// Позиционирование "CT".
func (r *Row) Underscore(text string, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, "CT")
	opts.FontSize = coalesce(opts.FontSize, r.core.DefaultFontSize().Sub(1))

	r.textCell(text, opts)

	return r
}

// UnderscoreSpan -- Underscore с заданным Colspan.
func (r *Row) UnderscoreSpan(text string, colspan uint8, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, "CT")
	opts.FontSize = coalesce(opts.FontSize, r.core.DefaultFontSize().Sub(1))
	opts.Colspan = colspan

	r.textCell(text, opts)

	return r
}

// Outlined -- ячейка таблицы с заданным текстом и рамкой из тонкой линии вокруг ячейки.
func (r *Row) Outlined(text string, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Border = coalesce(opts.Border, "o")

	r.textCell(text, opts)

	return r
}

// Skip -- пустая ячейка таблицы.
func (r *Row) Skip() *Row {
	r.skipCell()

	return r
}

// Image -- ячейка таблицы с заданным изображением. Изображение должно быть заранее задано методом [Core.ReadImage] или [Core.AddImage].
// Чтобы выбрать заданное изображение, необходимо указать его псевдоним. Если изображение не будет найдено, вместо него будет использован
// текст, записанный в [CellOptions.PlaceHolder].
// Размер изображения будет масштабирован по меньшей стороне ячейки.
func (r *Row) Image(alias string, options ...CellOptions) *Row {
	opts := getOptions(options)

	r.imageCell(alias, opts)

	return r
}

func (r *Row) PagesCount(options ...CellOptions) *Row {
	opts := getOptions(options)

	r.pageCountCell(opts)

	return r
}

func (r *Row) cell() *cell {
	for r.columnIndex+1 < r.columnsLen && r.rowspans[r.columnIndex] > 0 {
		r.columnIndex++
	}

	return &r.cells[r.columnIndex]
}

func (r *Row) skipCell() *cell {
	c := r.cell()

	c.colspan = 1
	c.rowspan = 1
	c.width = r.cellWidth(c.colspan)

	r.updateParameters(c)

	return c
}

func (r *Row) textCell(text string, options CellOptions) *cell {
	_font, err := r.core.font(coalesce(options.Font, r.core.fontRegularAlias))
	if err != nil {
		r.ctx.SetError(err)

		return nil
	}

	c := r.cell()

	c.mode = textMode
	c.font = _font
	c.id = options.ID
	c.height = options.Height
	c.textColor = options.TextColor
	c.borderColor = options.BorderColor
	c.wrapped = options.Wrap
	c.underline = options.UnderLine
	c.border = parseBorder(options.Border)
	c.borderSize = coalesce(options.BorderSize, r.core.DefaultBorderSize())
	c.alignH, c.alignV = parseAlignment(options.Align)
	c.fontSize = coalesce(options.FontSize, r.core.DefaultFontSize())
	c.colspan = coalesce(options.Colspan, 1)
	c.rowspan = coalesce(options.Rowspan, 1)
	c.width = r.cellWidth(c.colspan)

	text = coalesce(text, options.PlaceHolder)

	c.setTextLines(r.ctx, text)

	r.updateParameters(c)

	return c
}

func (r *Row) imageCell(alias string, options CellOptions) *cell {
	_image, err := r.getImage(alias)
	if err != nil {
		if r.core.ignoreImageNotFound {
			return r.textCell(options.PlaceHolder, options)
		}

		r.ctx.SetError(err)

		return nil
	}

	c := r.cell()

	c.mode = imageMode
	c.image = _image
	c.imageScale = coalesce(options.Scale, 1)
	c.offsetH = options.OffsetH
	c.offsetV = options.OffsetV
	c.id = options.ID
	c.height = options.Height
	c.textColor = options.TextColor
	c.borderColor = options.BorderColor
	c.border = parseBorder(options.Border)
	c.borderSize = coalesce(options.BorderSize, r.core.DefaultBorderSize())
	c.alignH, c.alignV = parseAlignment(options.Align)
	c.colspan = coalesce(options.Colspan, 1)
	c.rowspan = coalesce(options.Rowspan, 1)
	c.width = r.cellWidth(c.colspan)

	r.updateParameters(c)

	return c
}

func (r *Row) pageCountCell(options CellOptions) *cell {
	c := r.textCell("", options)
	if c == nil {
		return nil
	}

	r.counter.register(c)

	c.mode = counterMode

	return c
}

func (r *Row) getImage(alias string) (*image.Image, error) {
	_image, ok := r.res.GetImage(alias)
	if ok {
		return _image, nil
	}

	_image, ok = r.core.resources.GetImage(alias)
	if ok {
		return _image, nil
	}

	err := errs.ErrNotFound(alias)

	return nil, err
}

// При модификации текста может измениться высота ячейки.
func (r *Row) modifyCell(c *cell, text string) {
	c.setTextLines(r.ctx, text)

	r.setHeight(c)
}

// Ширина ячейки равна сумме ширин всех колонок, которые она занимает, и расстояний между этими колонками.
func (r *Row) cellWidth(colspan uint8) (w unit.MM) {
	for i := r.columnIndex; i < r.columnsLen && i < r.columnIndex+colspan; i++ {
		w += r.columns[i] + r.spacing
	}

	w -= r.spacing

	return w
}

// Поле cell.height должно быть равно высоте строки, но пока все ячейки не будут созданы, мы не знаем итоговую высоту строки.
// Поэтому высота ячейки будет определяться в методе render().
// TODO: сейчас я не считаю высоту, если есть rowspan, т.к. не знаю высоту следующей строки. Может быть нужно её учитывать
func (r *Row) setHeight(c *cell) {
	linesCount := coalesce(len(c.textLines), 1)

	if c.font != nil {
		calcHeight := c.font.Height(c.fontSize).MM() * unit.MM(linesCount)

		if calcHeight > r.height && c.rowspan < 2 {
			r.height = calcHeight
		}
	}

	if c.height > r.height {
		r.height = c.height
	}
}

// Если ячейки предыдущей строки имели rowspan, то мы ищем первую ячейку, которая rowspan не имела, и сдвигаем курсор на неё.
func (r *Row) setColIndex() {
	for r.columnIndex+1 < r.columnsLen && r.rowspans[r.columnIndex] > 0 {
		r.columnIndex++
	}
}

// Каждая ячейка имеет свой colspan > 0. Если colspan > 1, то пропущенным ячейкам тоже необходимо присвоить rowspan этой ячейки.
// Сдвигаем курсор к следующей ячейке.
func (r *Row) updateParameters(c *cell) {
	shift := r.columnIndex + c.colspan

	for i := r.columnIndex; i < shift && i < r.columnsLen; i++ {
		r.rowspans[i] += c.rowspan
		r.columnIndex = i
	}

	r.setHeight(c)
}

type cell struct {
	font        *font.Font
	image       *image.Image
	textLines   []font.Text
	width       unit.MM
	height      unit.MM
	offsetH     unit.MM
	offsetV     unit.MM
	fontSize    unit.PT
	borderSize  unit.PT
	imageScale  float64
	textColor   Color
	borderColor Color
	id          uint8
	border      uint8
	colspan     uint8
	rowspan     uint8
	alignH      uint8
	alignV      uint8
	mode        mode
	underline   bool
	wrapped     bool
}

func (c *cell) setTextLines(ctx *manager.Context, text string) {
	if c.font == nil {
		return
	}

	if c.textLines == nil {
		c.textLines = make([]font.Text, 0, 10)
	} else {
		c.textLines = c.textLines[:0]
	}

	if text == "" {
		return
	}

	width := unit.MM(0)
	if c.wrapped {
		width = c.width
	}

	c.font.SplitText(text, c.fontSize, width, &c.textLines, ctx.RuneBuffer(), ctx.HexBuffer())
}

func (c *cell) render(dst *stream.Stream, cursor unit.Point, parentTextColor, parentBorderColor Color) {
	if c.border > 0 {
		c.renderBorder(dst, cursor, parentBorderColor)
	}

	switch c.mode {
	case imageMode:
		c.renderImage(dst, cursor)
	case counterMode:
		c.renderXObject(dst, cursor, pages.AliasPagesCount)
	case textMode:
		c.renderText(dst, cursor, parentTextColor)
	default:
	}
}

func (c *cell) renderImage(dst *stream.Stream, cursor unit.Point) {
	var w, h, dx, dy unit.MM

	if c.height < c.width {
		h = c.imageHeight(0)
		w = c.imageWidth(h)
		dy = c.imageDy(h)
		dx = c.imageDx(w)
	} else {
		w = c.imageWidth(0)
		h = c.imageHeight(w)
		dx = c.imageDx(w)
		dy = c.imageDy(h)
	}

	matrix := primitives.NewMatrix(
		primitives.NewScale(w.PT(), h.PT()),
		primitives.Skew{},
		primitives.NewPoint(cursor.X().Add(dx).PT(), cursor.Y().Add(dy).PT().Abs()),
	)

	c.image.WriteToStream(dst, matrix)
}

func (c *cell) renderText(dst *stream.Stream, cursor unit.Point, parentTextColor Color) {
	if !c.textColor.isDefault() {
		c.textColor.WriteToStream(dst, primitives.FillColorRGB)
		defer parentTextColor.WriteToStream(dst, primitives.FillColorRGB)

		if c.underline {
			c.textColor.WriteToStream(dst, primitives.StrokeColorRGB)
			defer parentTextColor.WriteToStream(dst, primitives.StrokeColorRGB)
		}
	}

	for i := range c.textLines {
		dx := c.lineDx(i)
		dy := c.lineDy(i)

		c.textLines[i].SetMatrix(
			primitives.NewPoint(cursor.X().Add(dx).PT(), cursor.Y().Add(dy).PT().Abs()),
		)

		if c.underline {
			underlineDy := c.borderSize.MM()
			borderSize := font.UnderlineSize(c.fontSize)
			borderCursor := cursor

			borderCursor.AddX(dx)
			borderCursor.AddY(dy)

			renderBorder(dst, borderCursor, c.textLines[i].Width(), underlineDy, borderBottom, borderSize)
		}
	}

	c.font.WriteToStream(dst, c.textLines, c.fontSize)
}

func (c *cell) renderBorder(dst *stream.Stream, cursor unit.Point, parentBorderColor Color) {
	if !c.borderColor.isDefault() {
		c.borderColor.WriteToStream(dst, primitives.StrokeColorRGB)
		defer parentBorderColor.WriteToStream(dst, primitives.StrokeColorRGB)
	}

	renderBorder(dst, cursor, c.width, c.height, c.border, c.borderSize)
}

func (c *cell) renderXObject(dst *stream.Stream, cursor unit.Point, alias string) {
	matrix := primitives.NewMatrix(
		primitives.NewScale(1, 1),
		primitives.Skew{},
		primitives.NewPoint(cursor.X().PT(), cursor.Y().Add(c.height).PT().Abs()),
	)

	xObject := primitives.NewXObject(primitives.Alias(alias), matrix)

	xObject.WriteToStream(dst)
}

func (c *cell) lineDx(index int) (dx unit.MM) {
	textWidth := c.textLines[index].Width()

	switch c.alignH {
	case alignR:
		dx = c.width - textWidth - font.Padding(c.fontSize).MM()
	case alignC:
		dx = (c.width - textWidth) / 2
	default:
		dx = font.Padding(c.fontSize).MM()
	}

	return dx
}

func (c *cell) lineDy(index int) (dy unit.MM) {
	fontHeight := c.font.Height(c.fontSize).MM()
	lenTextLines := len(c.textLines)
	padding := font.Padding(c.fontSize).MM()

	// index + 1 необходим для того, чтобы выставить Y-координату по верхнему краю шрифта.
	// PDF считает Y от нижней границы страницы, а здесь все координаты указаны от верхней. К тому же позиционирует шрифт по baseline.
	// Поэтому для верного позиционирования шрифта нам необходимо добавить еще одну высоту строки.
	dy = unit.MM(index+1)*fontHeight - padding

	switch c.alignV {
	case alignB:
		dy += c.height - unit.MM(lenTextLines)*fontHeight
	case alignM:
		dy += (c.height - unit.MM(lenTextLines)*fontHeight) / 2
	default:
		dy += 0
	}

	return dy
}

func (c *cell) imageDx(width unit.MM) (dx unit.MM) {
	dx = c.offsetH

	switch c.alignH {
	case alignR:
		dx += c.width - width
	case alignC:
		dx += (c.width - width) / 2
	default:
		dx += 0
	}

	return dx
}

func (c *cell) imageDy(height unit.MM) (dy unit.MM) {
	dy = height + c.offsetV

	switch c.alignV {
	case alignB:
		dy += c.height - height
	case alignM:
		dy += (c.height - height) / 2
	default:
		dy += 0
	}

	return dy
}

func (c *cell) imageHeight(width unit.MM) (h unit.MM) {
	if width == 0 {
		return c.height * unit.MM(c.imageScale)
	}

	imgW := c.image.Width()
	scale := width / unit.MM(imgW)

	imgH := c.image.Height()
	h = unit.MM(imgH) * scale

	return h
}

func (c *cell) imageWidth(height unit.MM) (w unit.MM) {
	if height == 0 {
		return c.width * unit.MM(c.imageScale)
	}

	imgH := c.image.Height()
	scale := height / unit.MM(imgH)

	imgW := c.image.Width()
	w = unit.MM(imgW) * scale

	return w
}
