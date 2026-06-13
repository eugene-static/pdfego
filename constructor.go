package pdfego

import (
	"errors"
	"strconv"

	"github.com/eugene-static/pdfego/internal/buffer"
	"github.com/eugene-static/pdfego/internal/font"
	"github.com/eugene-static/pdfego/internal/image"
	"github.com/eugene-static/pdfego/unit"
)

const (
	defaultCell uint8 = iota
	imageCell
)

// Columns возвращает слайс значений в миллиметрах. Функция удобна для задания ширин колонок таблицы.
func Columns(columns ...unit.MM) []unit.MM {
	return columns
}

// Constructor -- главный узел дерева построения макета документа.
type Constructor struct {
	core      *Core
	paginator *Paginator
	block     Block
	x0        unit.MM
	y0        unit.MM
	x         unit.MM
	y         unit.MM
	dx        unit.MM
	dy        unit.MM
}

// NewConstructor создает новый экземпляр конструктора с заданным ядром. Задает первую страницу макета.
func NewConstructor(core *Core) *Constructor {
	core.startDocument()
	core.newPage()
	x0, y0 := core.page.x0y0()

	block := Block{
		core: core,
	}

	return &Constructor{
		core:  core,
		block: block,
		x0:    x0,
		y0:    y0,
		x:     x0,
		y:     y0,
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
func (c *Constructor) Build(applier BlockApplier) *Constructor {
	c.render()

	c.block.slots = c.block.slots[:0]
	applier(&c.block)

	return c
}

// NewPage создает новую страницу.
func (c *Constructor) NewPage() {
	c.render()
	c.renderPage()

	c.core.newPage()

	c.x, c.y = c.x0, c.y0
}

// Bytes завершает документ и отдает бинарные данные готового документа и ошибку, если такова была во время создания и рендеринга макета.
func (c *Constructor) Bytes() ([]byte, error) {
	c.render()
	c.renderPage()

	defer c.reset()

	c.core.finishDocument()

	defer c.core.reset()

	err := c.core.err()
	if err != nil {
		return nil, err
	}

	return c.core.bytes(), nil
}

// Возвращает экземпляр Block. Все блоки располагаются вертикально друг за другом.
// При каждом новом вызове метода, предыдущий экземпляр рендерится и больше не может быть изменен.
func (c *Constructor) Block(options ...NodeOptions) *Block {
	c.newBlock(options)

	return &c.block
}

// Возвращает экземпляр Header. Этот блок будет повторяться в начале каждой новой страницы с момента инициализации
// и до вызова метода ReleaseHeader(). Так же он отрисуется в момент инициализации.
func (c *Constructor) Header(options ...NodeOptions) *Header {
	c.newBlock(options)

	x0, y0 := c.core.page.x0y0()

	return &Header{
		block: &c.block,
		buf:   c.core.page.newHeader(),
		x0:    x0,
		y0:    y0,
		setHeight: func(height unit.MM) {
			c.y0 += height
		},
	}
}

// ReleaseHeader останавливает повторение блока Header на каждой странице.
func (c *Constructor) ReleaseHeader() {
	c.core.page.releaseHeader()

	c.y0 = c.core.page.y0()
}

// Инициализирует повторяющийся блок.
func (c *Constructor) Repeater(sectioner Sectioner, options ...NodeOptions) *Repeater {
	c.newBlock(options)

	return &Repeater{
		block:      &c.block,
		sectioner:  sectioner,
		renderFunc: c.render,
	}
}

// Возвращает экземпляр Watermark. Этот блок рендерится на каждой странице с момента инициализации.
// В отличие от Header не влияет на расположение других узлов конструктора.
func (c *Constructor) Watermark(options ...WatermarkOptions) *Watermark {
	opts := getOptions(options)
	alignH, alignV := parseAlignment(opts.Align)
	x0, y0 := c.core.page.x0y0()

	return &Watermark{
		block: &Block{
			core: c.core,
		},
		buf:    c.core.page.newWatermark(),
		x0:     x0,
		y0:     y0,
		alignH: alignH,
		alignV: alignV,
	}
}

// Возвращает экземпляр Paginator. Этот блок рендерится на каждой странице с момента инициализации.
// Располагается внизу документа в области между нижней границей и нижнем краем страницы.
// Пишет номер текущей страницы в первую ячейку с ID = 0.
func (c *Constructor) Paginator(options ...PaginatorOptions) *Paginator {
	opts := getOptions(options)
	x0 := c.core.page.x0()
	y0 := c.core.page.yB()
	alignH, alignV := parseAlignment(opts.Align)

	paginator := &Paginator{
		block:  &Block{core: c.core},
		buf:    c.core.page.buffer,
		x0:     x0,
		y0:     y0,
		offset: opts.Offset,
		alignH: alignH,
		alignV: alignV,
	}

	return paginator
}

// Paginate создает блок для отображения номера страницы.
func (c *Constructor) Paginate(offset int) {
	c.paginator = c.Paginator(PaginatorOptions{Offset: offset, Align: "CT"})

	c.paginator.Slot().Table(1, Columns(5)).Row().Cell("0", CellOptions{TextColor: ColorGray50})
}

func (c *Constructor) reset() {
	c.block.reset()

	c.x0, c.y0 = c.core.page.x0y0()
	c.x, c.y = c.x0, c.y0

	c.core.page.reset()
	c.core.startDocument()
	c.core.newPage()
}

// Новый блок не создается. Вся информация, записанная в блок рендерится и записывается в буфер страницы.
// После блок сбрасывается и к нему применяются новые опции. Поэтому нельзя изменять "старый блок" после вызова "нового".
func (c *Constructor) newBlock(options []NodeOptions) {
	opts := getOptions(options)

	c.render()

	c.block.slots = c.block.slots[:0]

	c.block.indentH = opts.IndentH
	c.block.indentV = opts.IndentV
	c.block.ledge = opts.Ledge
	c.block.spacing = opts.Spacing
	c.block.border = parseBorder(opts.Border)
	c.block.borderSize = opts.BorderSize
}

func (c *Constructor) render() {
	c.dy = c.block.height() + c.block.indentV

	if c.core.page.isBelowBottomBorder(c.y + c.dy) {
		c.renderPage()
		c.core.newPage()

		c.x, c.y = c.x0, c.y0
	}

	c.block.render(c.core.page.buffer, c.x, c.y)

	c.y += c.dy
}

func (c *Constructor) renderPage() {
	if c.paginator != nil {
		c.paginator.render()
	}

	c.core.renderPage()
}

// Block -- экземпляр блока. Все блоки располагаются друг за другом вертикально, независимо от того, кем был создан экземпляр.
// Блок в составе конструктора неделим: если его высота больше оставшегося места на странице, он будет отображен на следующей странице.
type Block struct {
	core       *Core
	slots      []Slot
	indentH    unit.MM
	indentV    unit.MM
	ledge      unit.MM
	spacing    unit.MM
	border     uint8
	borderSize unit.PT
}

type BlockApplier func(block *Block)

// Apply вызывает функцию func(block *Block).
func (b *Block) Apply(applier BlockApplier) {
	applier(b)
}

// Возвращает экземпляр Slot.
func (b *Block) Slot(options ...NodeOptions) *Slot {
	opts := getOptions(options)

	s := Slot{
		core:       b.core,
		indentH:    opts.IndentH,
		indentV:    opts.IndentV,
		spacing:    opts.Spacing,
		ledge:      opts.Ledge,
		borderSize: opts.BorderSize,
		border:     parseBorder(opts.Border),
	}

	b.slots = append(b.slots, s)

	return &b.slots[len(b.slots)-1]
}

func (b *Block) render(buf *buffer.Buffer, x, y unit.MM) {
	x += b.indentH
	y += b.indentV

	if b.border > 0 {
		width := b.width()
		height := b.height()
		borderSize := coalesce(b.borderSize, b.core.defaultBorderSize())

		renderBorder(buf, x, y, width, height, b.border, borderSize)
	}

	for i := range b.slots {
		b.slots[i].render(buf, x, y)

		x += b.slots[i].width() + b.spacing
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
	b.slots = b.slots[:0]

	b.indentH = 0
	b.indentV = 0
	b.ledge = 0
	b.spacing = 0
	b.border = 0
	b.borderSize = 0
}

// Header -- блок, повторяющийся в начале каждой страницы. Он является частью макета документа, поэтому не может быть изменен
// после создания нового блока.
type Header struct {
	block     *Block
	buf       *buffer.Buffer
	x0        unit.MM
	y0        unit.MM
	setHeight func(h unit.MM)
}

// Apply вызывает переданную функцию и сохраняет результат отрисовки в собственном буфере.
func (h *Header) Apply(applier BlockApplier) {
	applier(h.block)

	height := h.block.height()

	h.setHeight(height)

	h.block.render(h.buf, h.x0, h.y0)
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
		r.block.reset()

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

// Watermark -- блок, повторяющейся на каждой странице. Водяной знак не является частью основного макета документа и будет располагаться
// поверх него. Позиционирование блока происходит в пределах границ страницы.
type Watermark struct {
	block  *Block
	buf    *buffer.Buffer
	x0     unit.MM
	y0     unit.MM
	alignH uint8
	alignV uint8
}

// Apply вызывает переданную функцию и сохраняет результат отрисовки в собственном буфере.
func (w *Watermark) Apply(applier BlockApplier) {
	applier(w.block)

	dx := w.dx()
	dy := w.dy()

	w.block.render(w.buf, w.x0+dx, w.y0+dy)
}

func (w *Watermark) dx() (dx unit.MM) {
	switch w.alignH {
	case alignR:
		dx = w.block.core.page.width - w.block.core.page.marginLeft - w.block.core.page.marginRight - w.block.width()
	case alignM:
		dx = (w.block.core.page.width - w.block.core.page.marginLeft - w.block.core.page.marginRight - w.block.width()) / 2
	default:
		dx = 0
	}

	return dx
}

func (w *Watermark) dy() (dy unit.MM) {
	switch w.alignV {
	case alignB:
		dy = w.block.core.page.height - w.block.core.page.marginTop - w.block.core.page.marginBottom - w.block.height()
	case alignM:
		dy = (w.block.core.page.height - w.block.core.page.marginTop - w.block.core.page.marginBottom - w.block.height()) / 2
	default:
		dy = 0
	}

	return dy
}

// Paginator -- блок, повторяющийся на каждой странице между ее нижней границей и ее нижним краем.
// Пагинатор не является частью основного макета документа. Макет блока не перестраивается при каждом повторении,
// а модифицируется в зависимости от текста номера страницы.
type Paginator struct {
	block  *Block
	buf    *buffer.Buffer
	x0     unit.MM
	y0     unit.MM
	offset int
	alignH uint8
	alignV uint8
}

// Создает экземпляр Slot аналогично такому же методу у Block.
func (p *Paginator) Slot(options ...NodeOptions) *Slot {
	return p.block.Slot(options...)
}

// Apply вызывает переданную функцию.
func (p *Paginator) Apply(applier BlockApplier) {
	applier(p.block)
}

func (p *Paginator) render() {
	num := strconv.Itoa(p.block.core.page.count + p.offset)

	dx := p.dx()
	dy := p.dy()

	p.block.modify([]string{num})
	p.block.render(p.buf, p.x0+dx, p.y0+dy)
}

func (p *Paginator) dx() (dx unit.MM) {
	switch p.alignH {
	case alignR:
		dx = p.block.core.page.width - p.block.core.page.marginLeft - p.block.core.page.marginRight - p.block.width()
	case alignM:
		dx = (p.block.core.page.width - p.block.core.page.marginLeft - p.block.core.page.marginRight - p.block.width()) / 2
	default:
		dx = 0
	}

	return dx
}

func (p *Paginator) dy() (dy unit.MM) {
	switch p.alignV {
	case alignB:
		dy += p.block.core.page.marginBottom - p.block.height()
	case alignM:
		dy += (p.block.core.page.marginBottom - p.block.height()) / 2
	default:
		dy += 0
	}

	return dy
}

// Slot представляет собой слот. Слоты располагаются друг за другом горизонтально. Может содержать либо дочерние блоки Block, либо таблицу Table.
// Высота родительского блока будет являться высотой самого высокого слота.
type Slot struct {
	core       *Core
	blocks     []Block
	table      *Table
	indentH    unit.MM
	indentV    unit.MM
	ledge      unit.MM
	spacing    unit.MM
	border     uint8
	borderSize unit.PT
}

type SlotApplier func(slot *Slot)

// Создает дочерний экземпляр блока Block. Если к слоту была добавлена таблица Table, то конструктор вернет ошибку, а блоки не будут отрисованы.
// Ширина слота будет являться шириной самого широкого блока. Высота слота будет суммой высот всех блоков.
func (s *Slot) Block(options ...NodeOptions) *Block {
	if s.table != nil {
		err := errors.New("слот уже содержит таблицу")

		s.core.setError(err)
	}

	opts := getOptions(options)

	b := Block{
		core:       s.core,
		indentH:    opts.IndentH,
		indentV:    opts.IndentV,
		ledge:      opts.Ledge,
		spacing:    opts.Spacing,
		border:     parseBorder(opts.Border),
		borderSize: opts.BorderSize,
	}

	s.blocks = append(s.blocks, b)

	return &s.blocks[len(s.blocks)-1]
}

// Создает дочерний экземпляр таблицы Table. Таблица у слота может быть только одна. При повторном вызове метода, таблица перезапишется на новую.
// Для инициализации таблицы необходимо указать количество строк и слайс ширин всех столбцов.
func (s *Slot) Table(rowsNum uint8, columns []unit.MM, options ...TableOptions) *Table {
	if len(s.blocks) > 0 {
		err := errors.New("слот уже содержит блоки")
		s.core.setError(err)
	}

	opts := getOptions(options)

	columnsNum := uint8(len(columns))

	t := &Table{
		core:        s.core,
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

	s.table = t

	return t
}

// Apply вызывает переданную функцию.
func (s *Slot) Apply(applier SlotApplier) *Slot {
	applier(s)

	return s
}

func (s *Slot) render(buf *buffer.Buffer, x, y unit.MM) {
	x += s.indentH
	y += s.indentV

	if s.border > 0 {
		width := s.width()
		height := s.height()
		borderSize := coalesce(s.borderSize, s.core.defaultBorderSize())

		renderBorder(buf, x, y, width, height, s.border, borderSize)
	}

	if s.table != nil {
		s.table.render(buf, x, y)

		return
	}

	for i := range s.blocks {
		s.blocks[i].render(buf, x, y)

		y += s.blocks[i].height() + s.blocks[i].indentV + s.spacing
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

	t.setRowIndex()

	r := &t.rows[t.rowIndex]

	t.decrementRowSpans()
	t.newRow(r, opts)
	t.updateIndexes()

	return r
}

func (t *Table) newRow(row *Row, options RowOptions) {
	columnsLen := len(t.columns)
	start := int(t.rowIndex) * columnsLen
	end := start + columnsLen

	row.height = options.MinHeight
	row.core = t.core
	row.columns = t.columns
	row.columnsLen = uint8(columnsLen)
	row.cells = t.cellsPool[start:end]
	row.rowspans = t.rowspans
	row.spacing = t.spacingH
}

func (t *Table) render(buf *buffer.Buffer, x, y unit.MM) {
	if t.core.err() != nil {
		return
	}

	x += t.indentH
	y += t.indentV

	if !t.textColor.isDefault() {
		buf.WriteTextColor(t.textColor.rgb())
		defer buf.WriteTextColor(ColorDefault.rgb())
	}

	if t.border > 0 {
		if !t.borderColor.isDefault() {
			buf.WriteBorderColor(t.borderColor.rgb())
			defer buf.WriteBorderColor(ColorDefault.rgb())
		}

		width := t.width()
		height := t.height()
		borderSize := coalesce(t.borderSize, t.core.defaultBorderSize())

		renderBorder(buf, x, y, width, height, t.border, borderSize)
	}

	cx := x
	cy := y

	for i := range t.cellsPool {
		ri := i / len(t.columns)
		ci := i % len(t.columns)

		if t.cellsPool[i].busy {
			t.setCellHeight(ri, i)

			t.cellsPool[i].render(buf, cx, cy, t.textColor, t.borderColor)
		}

		cx += t.columns[ci] + t.spacingH

		if ci == len(t.columns)-1 {
			cx = x
			cy += t.rows[ri].height + t.spacingV
		}
	}
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

// Нельзя допустить вызов строки больше, чем задано при инициализации таблицы.
func (t *Table) setRowIndex() {
	if int(t.rowIndex) >= len(t.rows) {
		t.rowIndex--
	}
}

func (t *Table) updateIndexes() {
	t.rowIndex++
}

// При создании новой строки уменьшаем все rowspan на 1.
func (t *Table) decrementRowSpans() {
	for i := range t.rowspans {
		if t.rowspans[i] > 0 {
			t.rowspans[i]--
		}
	}
}

// Row представляет собой экземпляр строки таблицы Table. На расстояние между строками влияет значение TableOptions.SpacingV.
type Row struct {
	core        *Core
	cells       []cell
	columns     []unit.MM
	rowspans    []uint8
	height      unit.MM
	spacing     unit.MM
	columnIndex uint8
	columnsLen  uint8
}

// Cell -- ячейка таблицы с заданным текстом. Использует шрифт, цвет по-умолчанию. Без рамки, позиционирование CM.
func (r *Row) Cell(text string, options ...CellOptions) *Row {
	opts := getOptions(options)

	r.newCell(text, defaultCell, opts)

	return r
}

// Label -- ячейка таблицы с заданным текстом. Без рамки, позиционирование LB.
func (r *Row) Label(text string, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, "LB")

	r.newCell(text, defaultCell, opts)

	return r
}

// LabelSpan -- Label с заданным Colspan.
func (r *Row) LabelSpan(text string, colspan uint8, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, "LB")
	opts.Colspan = colspan

	r.newCell(text, defaultCell, opts)

	return r
}

// LabelHead -- Label, где в качестве шрифта используется заданный шрифт с псевдонимом "BOLD".
func (r *Row) LabelHead(text string, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, "LB")
	opts.Font = coalesce(opts.Font, FontBold)

	r.newCell(text, defaultCell, opts)

	return r
}

// Blank -- ячейка таблицы, бланк, с заданным текстом и позиционированием. Тонкая граница снизу по-умолчанию.
func (r *Row) Blank(text, align string, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, align, "CB")
	opts.Border = coalesce(opts.Border, "b")

	r.newCell(text, defaultCell, opts)

	return r
}

// BlankSpan -- Blank с заданным Colspan.
func (r *Row) BlankSpan(text, align string, colspan uint8, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, align, "CB")
	opts.Border = coalesce(opts.Border, "b")
	opts.Colspan = colspan

	r.newCell(text, defaultCell, opts)

	return r
}

// BlankEmpty -- пустой Blank, без текста.
func (r *Row) BlankEmpty(options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Border = coalesce(opts.Border, "b")

	r.newCell("", defaultCell, opts)

	return r
}

// Form -- Blank с переносом текста.
func (r *Row) Form(text, align string, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, align, "CB")
	opts.Border = coalesce(opts.Border, "b")
	opts.Wrap = true

	r.newCell(text, defaultCell, opts)

	return r
}

// FormSpan -- BlankSpan с переносом текста.
func (r *Row) FormSpan(text, align string, colspan uint8, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, align, "CB")
	opts.Border = coalesce(opts.Border, "b")
	opts.Colspan = colspan
	opts.Wrap = true

	r.newCell(text, defaultCell, opts)

	return r
}

// Paragraph -- ячейка таблицы с заданным текстом. Позиционирование "CB"
func (r *Row) Paragraph(text string, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, "CB")

	r.newCell(text, defaultCell, opts)

	return r
}

// Underscore -- ячейка таблицы с заданным текстом. Размер шрифта меньше заданного по-умолчанию на 1 пункт.
// Позиционирование "CT".
func (r *Row) Underscore(text string, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, "CT")
	opts.FontSize = coalesce(opts.FontSize, r.core.defaultFontSize().Sub(1))

	r.newCell(text, defaultCell, opts)

	return r
}

// UnderscoreSpan -- Underscore с заданным Colspan.
func (r *Row) UnderscoreSpan(text string, colspan uint8, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Align = coalesce(opts.Align, "CT")
	opts.FontSize = coalesce(opts.FontSize, r.core.defaultFontSize().Sub(1))
	opts.Colspan = colspan

	r.newCell(text, defaultCell, opts)

	return r
}

// Outlined -- ячейка таблицы с заданным текстом и рамкой из тонкой линии вокруг ячейки.
func (r *Row) Outlined(text string, options ...CellOptions) *Row {
	opts := getOptions(options)

	opts.Border = coalesce(opts.Border, "o")

	r.newCell(text, defaultCell, opts)

	return r
}

// Skip -- пустая ячейка таблицы.
func (r *Row) Skip() *Row {
	r.newCell("", defaultCell, CellOptions{})

	return r
}

// Image -- ячейка таблицы с заданным изображением. Изображение должно быть заранее задано методом [Core.ReadImage] или [Core.AddImage].
// Чтобы выбрать заданное изображение, необходимо указать его псевдоним. Если изображение не будет найдено, вместо него будет использован
// текст, записанный в [CellOptions.PlaceHolder].
// Размер изображения будет масштабирован по меньшей стороне ячейки.
func (r *Row) Image(alias string, options ...CellOptions) *Row {
	opts := getOptions(options)

	r.newCell(alias, imageCell, opts)

	return r
}

func (r *Row) newCell(text string, profile uint8, options CellOptions) {
	r.setColIndex()

	c := &r.cells[r.columnIndex]

	c.id = options.ID
	c.height = options.Height
	c.static = options.Height > 0
	c.offsetH = options.OffsetH
	c.offsetV = options.OffsetV
	c.textColor = options.TextColor
	c.borderColor = options.BorderColor
	c.border = parseBorder(options.Border)
	c.borderSize = coalesce(options.BorderSize, r.core.defaultBorderSize())
	c.colspan = coalesce(options.Colspan, 1)
	c.rowspan = coalesce(options.Rowspan, 1)
	c.width = r.cellWidth(c.colspan)
	c.busy = true

	if profile == imageCell {
		c.image = r.core.image(text)
		c.imageScale = coalesce(options.Scale, 1)

		text = options.PlaceHolder
	}

	text = coalesce(text, options.PlaceHolder)

	if text != "" {
		c.wrapped = options.Wrap
		c.alignH, c.alignV = parseAlignment(options.Align)
		c.fontSize = coalesce(options.FontSize, r.core.defaultFontSize())
		c.font = r.core.font(coalesce(options.Font, FontRegular))
		c.textLines = make([]font.Text, 0, 10)

		c.setTextLines(text)
	}

	r.setHeight(c)
	r.updateIndexes(c)
}

// При модификации текста может измениться высота ячейки.
func (r *Row) modifyCell(c *cell, text string) {
	c.setTextLines(text)

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
	if c.font != nil {
		calcHeight := c.font.Height(c.fontSize).MM() * unit.MM(len(c.textLines))

		if calcHeight > r.height && c.rowspan < 2 {
			r.height = calcHeight
		}
	}

	if c.static && c.height > r.height {
		r.height = c.height
	}
}

// Если ячейки предыдущей строки имели rowspan, то мы ищем первую ячейку, которая rowspan не имела, и сдвигаем курсор на неё.
func (r *Row) setColIndex() {
	for r.columnIndex < r.columnsLen && r.rowspans[r.columnIndex] > 0 {
		r.columnIndex++
	}
}

// Каждая ячейка имеет свой colspan > 0. Если colspan > 1, то пропущенным ячейкам тоже необходимо присвоить rowspan этой ячейки.
// Сдвигаем курсор к следующей ячейке.
func (r *Row) updateIndexes(c *cell) {
	shift := r.columnIndex + c.colspan

	for i := r.columnIndex; i < min(shift, r.columnsLen); i++ {
		r.rowspans[i] += c.rowspan
		r.columnIndex = i
	}
}

type cell struct {
	id          uint8
	font        *font.Font
	image       *image.Image
	textColor   Color
	borderColor Color
	textLines   []font.Text
	imageScale  float64
	width       unit.MM
	height      unit.MM
	offsetH     unit.MM
	offsetV     unit.MM
	fontSize    unit.PT
	borderSize  unit.PT
	border      uint8
	colspan     uint8
	rowspan     uint8
	alignH      uint8
	alignV      uint8
	wrapped     bool
	static      bool
	busy        bool
}

func (c *cell) setTextLines(text string) {
	if c.font == nil || text == "" {
		return
	}

	c.textLines = c.textLines[:0]

	if c.wrapped {
		c.textLines = append(c.textLines, c.font.SplitText(text, c.fontSize, c.width, c.textLines)...)

		return
	}

	c.textLines = append(c.textLines, c.font.FullText(text, c.fontSize))
}

func (c *cell) render(buf *buffer.Buffer, x, y unit.MM, parentTextColor, parentBorderColor Color) {
	if c.border > 0 {
		c.renderBorder(buf, x, y, parentBorderColor)
	}

	if c.image != nil {
		c.renderImage(buf, x, y)

		return
	}

	if len(c.textLines) != 0 {
		c.renderText(buf, x, y, parentTextColor)

		return
	}
}

func (c *cell) renderImage(buf *buffer.Buffer, x, y unit.MM) {
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

	buf.WriteImage(x+dx, y+dy, w, h, c.image.Alias())
}

// BT /[FontAlias] [FontSize] Tf 1 0 0 1 [X] [Y] Tm <[TextHex]> Tj ET
func (c *cell) renderText(buf *buffer.Buffer, x, y unit.MM, parentTextColor Color) {
	if !c.textColor.isDefault() {
		buf.WriteTextColor(c.textColor.rgb())
		defer buf.WriteTextColor(parentTextColor.rgb())
	}

	buf.WriteStringLn("BT")
	buf.WriteFont(c.font.Alias(), c.fontSize)

	for i := range c.textLines {
		dx := c.lineDx(i)
		dy := c.lineDy(i)

		buf.WriteText(c.font, x+dx, y+dy, c.textLines[i].Data())
	}

	buf.WriteStringLn("ET")
}

func (c *cell) renderBorder(buf *buffer.Buffer, x, y unit.MM, parentBorderColor Color) {
	if !c.borderColor.isDefault() {
		buf.WriteBorderColor(c.borderColor.rgb())
		defer buf.WriteBorderColor(parentBorderColor.rgb())
	}

	renderBorder(buf, x, y, c.width, c.height, c.border, c.borderSize)
}

func (c *cell) lineDx(index int) (dx unit.MM) {
	textWidth := c.textLines[index].Width()

	if textWidth == 0 {
		text := []rune(c.textLines[index].Data())
		textWidth = c.font.MeasureText(c.fontSize, text, 0, len(text)).MM()
	}

	switch c.alignH {
	case alignR:
		dx = c.width - textWidth - font.Margin(c.fontSize).MM()
	case alignC:
		dx = (c.width - textWidth) / 2
	default:
		dx = font.Margin(c.fontSize).MM()
	}

	return dx
}

func (c *cell) lineDy(index int) (dy unit.MM) {
	fontHeight := c.font.Height(c.fontSize).MM()
	lenTextLines := len(c.textLines)
	padding := font.Margin(c.fontSize).MM()

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
