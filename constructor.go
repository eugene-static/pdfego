package pdf_craft

import (
	"errors"

	"github.com/eugene-static/pdf-craft/internal/buffer"
	"github.com/eugene-static/pdf-craft/internal/font"
	"github.com/eugene-static/pdf-craft/pkg/unit"
)

const (
	blankBlock = iota
	defaultBlock
	headerBlock
	headerStopBlock
	repeatableBlock
)

type Constructor struct {
	core    *Core
	block   Block
	ordered Ordered
	x0      unit.MM
	y0      unit.MM
	x       unit.MM
	y       unit.MM
}

func New(core *Core) *Constructor {
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

func (c *Constructor) NewPage() {
	c.newBlock(blankBlock, nil)
	c.core.renderPage()
	c.core.newPage()

	c.x, c.y = c.x0, c.y0
}

func (c *Constructor) Render() {
	c.render()
	c.core.finishDocument()
}

type Ordered interface {
	Len() int
	OrderedRow(int) []string
}

func (c *Constructor) Bytes() ([]byte, error) {
	return c.core.bytes(), c.core.err()
}

func (c *Constructor) Block(opts ...NodeOptions) *Block {
	c.newBlock(defaultBlock, opts)

	return &c.block
}

func (c *Constructor) Header(opts ...NodeOptions) *Block {
	c.newBlock(headerBlock, opts)

	return &c.block
}

func (c *Constructor) EndHeader(opts ...NodeOptions) *Block {
	c.newBlock(headerStopBlock, opts)

	return &c.block
}

func (c *Constructor) Repeater(ordered Ordered, opts ...NodeOptions) *Constructor {
	c.ordered = ordered
	c.newBlock(repeatableBlock, opts)

	return c
}

// TODO:
func (c *Constructor) AddRepeater(filler RepeaterFiller) {
	if c.ordered.Len() == 0 {
		return
	}

	ordered := c.ordered.OrderedRow(0)

	filler(&c.block, ordered)

	for i := range c.ordered.Len() - 1 {
		c.render()

		ordered = c.ordered.OrderedRow(i + 1)

		c.block.update(ordered)
	}
}

func (c *Constructor) newBlock(profile byte, opts []NodeOptions) *Block {
	options := getOptions(opts)

	c.render()

	c.block.slots = c.block.slots[:0]
	c.block.profile = profile
	c.block.opts = options

	return &c.block
}

func (c *Constructor) render() {
	height := c.block.height()

	if c.block.profile == headerBlock {
		header := c.core.page.newHeader()

		c.block.render(header, c.x0, c.y0)
		c.y0 += height
	}

	if c.block.profile == headerStopBlock {
		c.core.page.removeHeader()

		_, c.y0 = c.core.page.x0y0()
	}

	if c.core.page.isBelowBottomBorder(c.y + c.block.opts.IndentTop + height) {
		c.core.renderPage()
		c.core.newPage()

		c.x, c.y = c.x0, c.y0
	}

	c.block.render(c.core.page.buffer, c.x, c.y)

	c.y += height
}

type Block struct {
	profile byte
	core    *Core
	slots   []Slot
	opts    NodeOptions
}

type BlockFiller func(*Block)

type RepeaterFiller func(*Block, []string)

func (b *Block) Slot(opts ...NodeOptions) *Slot {
	options := getOptions(opts)

	s := Slot{
		core: b.core,
		opts: options,
	}

	b.slots = append(b.slots, s)

	return &b.slots[len(b.slots)-1]
}

func (b *Block) Add(filler BlockFiller) *Block {
	filler(b)

	return b
}

func (b *Block) render(buf *buffer.Buffer, x, y unit.MM) {
	x += b.opts.IndentLeft
	y += b.opts.IndentTop

	for i := range b.slots {
		b.slots[i].render(buf, x, y)

		x += b.slots[i].width() + b.opts.Spacing
	}
}

func (b *Block) update(ordered []string) {
	for i := range b.slots {
		b.slots[i].update(ordered)
	}
}

func (b *Block) height() (h unit.MM) {
	for i := range b.slots {
		sh := b.slots[i].height()
		if h < sh {
			h = sh
		}
	}

	return h + b.opts.Ledge
}

func (b *Block) width() (w unit.MM) {
	for i := range b.slots {
		w += b.slots[i].width() + b.slots[i].opts.IndentLeft
	}

	w += b.opts.Spacing * unit.MM(len(b.slots)-1)

	return w
}

type Slot struct {
	core   *Core
	blocks []Block
	table  *Table
	opts   NodeOptions
}

type SlotFiller func(*Slot)

func (s *Slot) Block(opts ...NodeOptions) *Block {
	if s.table != nil {
		err := errors.New("слот уже содержит таблицу")
		s.core.writeError(err)
	}

	options := getOptions(opts)

	b := Block{
		core: s.core,
		opts: options,
	}

	s.blocks = append(s.blocks, b)

	return &s.blocks[len(s.blocks)-1]
}

func (s *Slot) Table(rowsNum int, columns []unit.MM) *Table {
	if len(s.blocks) > 0 {
		err := errors.New("слот уже содержит блоки")
		s.core.writeError(err)
	}

	columnsNum := len(columns)

	t := &Table{
		core:      s.core,
		columns:   columns,
		rows:      make([]Row, rowsNum),
		rowspans:  make([]uint8, columnsNum),
		cellsPool: make([]cell, rowsNum*columnsNum),
	}

	s.table = t

	return t
}

func (s *Slot) Add(filler SlotFiller) *Slot {
	filler(s)

	return s
}

func (s *Slot) render(buf *buffer.Buffer, x, y unit.MM) {
	x += s.opts.IndentLeft
	y += s.opts.IndentTop

	if s.table != nil {
		s.table.render(buf, x, y)

		return
	}

	if s.opts.Border != "" {
		borderMask := parseBorder(s.opts.Border)
		width := s.width()
		height := s.height() + s.opts.IndentTop

		borderSize := s.core.DefaultBorderSize()
		if s.opts.BorderSize > 0 {
			borderSize = s.opts.BorderSize
		}

		renderBorder(buf, x, y, width, height, borderMask, borderSize)
	}

	for i := range s.blocks {
		s.blocks[i].render(buf, x, y)

		y += s.blocks[i].height()
	}
}

func (s *Slot) update(ordered []string) {
	if s.table != nil {
		s.table.update(ordered)
	}

	for i := range s.blocks {
		s.blocks[i].update(ordered)
	}
}

func (s *Slot) height() (h unit.MM) {
	h = s.opts.Ledge

	if len(s.blocks) > 0 {
		for i := range s.blocks {
			h += s.blocks[i].height() + s.blocks[i].opts.IndentTop + s.blocks[i].opts.Ledge
		}

		return h
	}

	h = s.table.height() + s.table.opts.IndentTop

	return h
}

func (s *Slot) width() (w unit.MM) {
	if s.table != nil {
		w = s.table.width()

		return w
	}

	for i := range s.blocks {
		bw := s.blocks[i].width() + s.blocks[i].opts.IndentLeft
		if w < bw {
			w = bw
		}
	}

	return w
}

type Table struct {
	core      *Core
	columns   []unit.MM
	rows      []Row
	cellsPool []cell
	rowspans  []uint8
	rowIndex  uint8
	opts      NodeOptions
}

func (t *Table) Row(opts ...NodeOptions) *Row {
	options := getOptions(opts)

	t.setRowIndex()

	r := &t.rows[t.rowIndex]

	t.decrementRowSpans()
	t.newRow(r, options)
	t.updateIndexes()

	return r
}

func (t *Table) Add(rowFunc func(t *Table)) {
	rowFunc(t)
}

func (t *Table) newRow(row *Row, options NodeOptions) {
	columnsLen := len(t.columns)
	start := int(t.rowIndex) * columnsLen
	end := start + columnsLen

	row.core = t.core
	row.columns = t.columns
	row.columnsLen = uint8(columnsLen)
	row.cells = t.cellsPool[start:end]
	row.rowspans = t.rowspans
	row.options = options
}

func (t *Table) render(buf *buffer.Buffer, x, y unit.MM) {
	cx := x
	cy := y

	for i := range t.cellsPool {
		ri := i / len(t.columns)
		ci := i % len(t.columns)

		if ci == 0 {
			cx += t.rows[ri].options.IndentLeft
			cy += t.rows[ri].options.IndentTop
		}

		if t.cellsPool[i].busy {
			t.setCellHeight(ri, i)

			t.cellsPool[i].render(buf, cx, cy)
		}

		cx += t.columns[ci]

		if ci == len(t.columns)-1 {
			cx = x
			cy += t.rows[ri].height + t.rows[ri].options.Ledge + t.opts.Spacing
		}
	}
}

func (t *Table) update(ordered []string) {
	for i := range t.cellsPool {
		ri := i / len(t.columns)
		ci := i % len(t.columns)

		if ci == 0 {
			t.rows[ri].height = 0
		}

		id := t.cellsPool[i].id
		if int(id) > len(ordered)-1 {
			continue
		}

		t.rows[ri].updateCell(&t.cellsPool[i], ordered[id])
	}
}

func (t *Table) height() (h unit.MM) {
	for i := range t.rows {
		h += t.rows[i].height + t.opts.Spacing
	}

	h += t.opts.Ledge - t.opts.Spacing

	return h
}

func (t *Table) width() (w unit.MM) {
	for i := range t.columns {
		w += t.columns[i]
	}

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

type Row struct {
	core        *Core
	cells       []cell
	columns     []unit.MM
	rowspans    []uint8
	height      unit.MM
	columnIndex uint8
	columnsLen  uint8
	options     NodeOptions
}

func (r *Row) Cell(text string, opts ...CellOptions) *Row {
	options := getOptions(opts)

	r.setColIndex()

	c := &r.cells[r.columnIndex]

	r.newCell(c, text, options)
	r.setHeight(c)
	r.updateIndexes(c)

	return r
}

func (r *Row) Label(text string) *Row {
	return r.Cell(text, CellOptions{Align: "LB"})
}

func (r *Row) LabelHead(text string) *Row {
	return r.Cell(text, CellOptions{Align: "LB", Font: FontBold})
}

func (r *Row) FormL(text string, wrapText bool) *Row {
	return r.Cell(text, CellOptions{Align: "LB", Border: "b", Wrap: wrapText})
}

func (r *Row) FormC(text string, wrapText bool) *Row {
	return r.Cell(text, CellOptions{Align: "CB", Border: "b", Wrap: wrapText})
}

func (r *Row) Paragraph(text string) *Row {
	return r.Cell(text, CellOptions{Align: "CB"})
}

func (r *Row) Underscore(text string) *Row {
	return r.Cell(text, CellOptions{Align: "CT", FontSize: r.core.DefaultFontSize().Sub(1)})
}

func (r *Row) Bounded(text string) *Row {
	return r.Cell(text, CellOptions{Align: "CM", Border: "o", Wrap: false})
}

func (r *Row) newCell(c *cell, text string, opts CellOptions) {
	c.textLines = make([]font.Text, 0, 10)
	c.font = r.core.font(FontRegular)
	c.fontSize = r.core.DefaultFontSize()
	c.borderSize = r.core.DefaultBorderSize()
	c.alignH = alignC
	c.alignV = alignM
	c.colspan = 1
	c.rowspan = 1
	c.busy = true

	c.setOpts(r.core, opts)

	c.width = r.cellWidth(c.colspan)

	c.setTextLines(text)
}

func (r *Row) updateCell(c *cell, text string) {
	c.setTextLines(text)

	r.setHeight(c)
}

func (r *Row) width() (w float64) {
	for i := range r.cells {
		w += r.cells[i].width.Float64()
	}

	return w
}

// Ширина ячейки равна сумме ширин всех колонок, которые она занимает.
func (r *Row) cellWidth(colspan uint8) (w unit.MM) {
	for i := r.columnIndex; i < r.columnsLen && i < r.columnIndex+colspan; i++ {
		w += r.columns[i]
	}

	return w
}

// Поле cell.height должно быть равно высоте строки, но пока все ячейки не будут созданы, мы не знаем итоговую высоту строки.
// Поэтому высота ячейки будет определяться в методе render().
// TODO: сейчас я не считаю высоту, если есть rowspan, т.к. не знаю высоту следующей строки. Может быть нужно её учитывать
func (r *Row) setHeight(c *cell) {
	calcHeight := c.font.Height(c.fontSize).MM() * unit.MM(len(c.textLines))

	if calcHeight > r.height && c.rowspan < 2 {
		r.height = calcHeight
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
	id         uint8
	font       *font.Font
	textLines  []font.Text
	width      unit.MM
	height     unit.MM
	fontSize   unit.PT
	borderSize unit.PT
	borderMask uint8
	colspan    uint8
	rowspan    uint8
	alignH     uint8
	alignV     uint8
	wrapped    bool
	static     bool
	busy       bool
}

func (c *cell) setOpts(core *Core, opts CellOptions) {
	c.id = opts.ID
	c.wrapped = opts.Wrap
	c.height = opts.Height
	c.static = opts.Height > 0
	c.borderMask = parseBorder(opts.Border)
	c.alignH, c.alignV = parseAlignment(opts.Align)

	if opts.Font != "" {
		c.font = core.font(opts.Font)
	}

	if opts.FontSize > 0 {
		c.fontSize = opts.FontSize
	}

	if opts.BorderSize > 0 {
		c.borderSize = opts.BorderSize
	}

	if opts.Colspan > 1 {
		c.colspan = opts.Colspan
	}

	if opts.Rowspan > 1 {
		c.rowspan = opts.Rowspan
	}
}

func (c *cell) setTextLines(text string) {
	c.textLines = c.textLines[:0]

	if text == "" {
		return
	}

	if c.wrapped {
		c.textLines = append(c.textLines, c.font.SplitText(text, c.fontSize, c.width, c.textLines)...)

		return
	}

	c.textLines = append(c.textLines, c.font.FullText(text, c.fontSize))
}

// BT /[FontAlias] [FontSize] Tf 1 0 0 1 [X] [Y] Tm <[TextHex]> Tj ET
func (c *cell) render(buf *buffer.Buffer, x, y unit.MM) {
	if c.borderMask != 0 {
		renderBorder(buf, x, y, c.width, c.height, c.borderMask, c.borderSize)
	}

	if len(c.textLines) == 0 {
		return
	}

	buf.WriteStringLn("BT")
	buf.WriteFont(c.font.Alias(), c.fontSize)

	for i := range c.textLines {
		dx := c.lineDx(i)
		dy := c.lineDy(i)

		cx := x + dx
		cy := y + dy

		buf.WriteText(c.font, cx, cy, c.textLines[i].Data())
	}

	buf.WriteStringLn("ET")
}

func (c *cell) lineDx(index int) (dx unit.MM) {
	textWidth := c.textLines[index].Width()

	if textWidth == 0 {
		text := []rune(c.textLines[index].Data())
		textWidth = c.font.MeasureText(c.fontSize, text, 0, len(text)).MM()
	}

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
