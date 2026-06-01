package _dev

//
//import (
//	"strconv"
//
//	"codeberg.org/go-pdf/fpdf"
//)
//
//const (
//	lineHeight          = 3
//	subLineHeight       = lineHeight - 1
//	margin              = 3
//	topMargin           = 3
//	mainFontSize        = 6
//	subFontSize         = mainFontSize - 1
//	separatorLeftMargin = 20
//	paragraphWidth      = 10
//	a4LandscapeWidth    = 297
//)
//
//type Template struct {
//	document *fpdf.Fpdf
//	table    table
//	entity   HTML
//	posX     float64
//	posY     float64
//}
//
//type table struct {
//	widths           map[string]float64
//	currentRowHeight float64
//	showHeader       bool
//}
//
//func New(entity HTML) Template {
//	pdf := fpdf.New("L", "mm", "A4", "")
//
//	tmpl := Template{
//		document: pdf,
//		entity:   entity,
//		table: table{
//			widths:           getTableWidths(),
//			currentRowHeight: lineHeight,
//			showHeader:       true,
//		},
//	}
//
//	pdf.SetCompression(false)
//	pdf.SetAutoPageBreak(true, margin)
//	pdf.AddUTF8Font("NotoSerif", "", "./fonts/NotoSerifSC-Regular.ttf")
//	pdf.AddUTF8Font("NotoSerif", "B", "./fonts/NotoSerifSC-ExtraBold.ttf")
//	pdf.SetFont("NotoSerif", "", mainFontSize)
//	pdf.SetMargins(margin, margin, -1)
//	pdf.NewPage()
//
//	tmpl.TitleHeader()
//	tmpl.Header()
//	tmpl.HeaderRequisites()
//	tmpl.TableHeader()
//	tmpl.TableLetterHeader()
//
//	tmpl.document.SetHeaderFunc(tmpl.TableLetterHeader)
//
//	for _, item := range entity.Details {
//		tmpl.TableBodyCells(item)
//	}
//
//	tmpl.TableFooterCells()
//	tmpl.Signatures()
//
//	return tmpl
//}
//
//func (t *Template) SavePosition() {
//	x, y := t.document.GetXY()
//
//	t.posX = x
//	t.posY = y
//}
//
//func (t *table) ResetRowHeight() {
//	t.currentRowHeight = lineHeight
//}
//
//func (t *Template) TitleHeader() {
//	t.document.MultiCell(separatorLeftMargin, 3, "Универсальный передаточный документ", "0", "L", false)
//	t.document.Cell(separatorLeftMargin, 5, "")
//	t.nl()
//	t.document.CellFormat(separatorLeftMargin-8, 5, "Статус", "0", 0, "L", false, 0, "")
//	t.document.SetLineWidth(0.3)
//	t.document.CellFormat(separatorLeftMargin-12, 5, "1", "1", 1, "C", false, 0, "")
//	t.document.SetLineWidth(0.1)
//
//	t.document.Cell(separatorLeftMargin, 5, "")
//	t.nl()
//	t.document.SetFontSize(subFontSize)
//	t.document.MultiCell(separatorLeftMargin, 2, "1 - счет-фактура и передаточный документ (акт)", "0", "L", false)
//	t.document.MultiCell(separatorLeftMargin, 2, "2 - передаточный документ (акт)", "0", "L", false)
//}
//
//func (t *Template) Header() {
//	t.document.SetFontSize(mainFontSize)
//	t.document.SetLeftMargin(separatorLeftMargin + margin + 5)
//	t.document.SetY(topMargin)
//
//	t.text(25, "Счет-фактура №")
//	t.formCenter(30, t.entity.SfNum)
//	t.textCenter(5, "от")
//	t.formCenter(30, "")
//	t.paragraph("(1)")
//
//	t.text(25, "Исправление №")
//	t.formCenter(30, "")
//	t.textCenter(5, "от")
//	t.formCenter(30, "")
//	t.paragraph("(1а)")
//
//	t.document.SetLeftMargin(100 + margin)
//	t.document.SetY(topMargin)
//
//	t.document.SetFontSize(subFontSize)
//	t.cellFormat(0, 2, "Приложение № 1 к постановлению Правительства Российской федерации от 26 декабря 2011 г. № 1137", "0", 1, "TR")
//	t.cellFormat(0, 2, "(в редакции постановления Правительства Российской Федерации от 16 августа Исправление № от (1а) 2024 г. № 1096)", "0", 1, "TR")
//	t.document.SetFontSize(mainFontSize)
//}
//
//func (t *Template) HeaderRequisites() {
//	const (
//		labelWidth = 50
//		formWidth  = 70
//	)
//
//	t.document.SetLeftMargin(separatorLeftMargin + margin + 5)
//	t.document.SetY(2*lineHeight + 5 + topMargin)
//
//	t.labelBoldFormMultiline(labelWidth, "Продавец:", formWidth, "", "(2)")
//	t.labelFormMultiline(labelWidth, "Адрес:", formWidth, "", "(2а)")
//	t.labelForm(labelWidth, "ИНН/КПП продавца:", formWidth, "", "(2б)")
//	t.labelForm(labelWidth, "Грузоотправитель и его адрес:", formWidth, "", "(3)")
//	t.labelFormMultiline(labelWidth, "Грузополучатель и его адрес:", formWidth, "", "(4)")
//	t.labelForm(labelWidth, "К платежно-расчетному документу №:", formWidth, "", "(5)")
//	t.labelFormMultiline(labelWidth, "Документ об отгрузке: наименование, №:", formWidth, "", "(5а)")
//
//	t.text(90, "К счету-фактуре (счетам-фактурам), выставленному (выставленным)")
//	t.nl()
//	t.text(90, "при получении оплаты, частичной оплаты или иных платежей в счет")
//	t.nl()
//	t.text(90, "предстоящих поставок товаров (выполнения работ, оказания услуг),")
//	t.nl()
//	t.text(labelWidth, "передачи имущественных прав №:")
//	t.form(32.5, "")
//	t.textCenter(5, "от")
//	t.form(32.5, "")
//	t.nl()
//	t.text(labelWidth, "исправление №:")
//	t.form(32.5, "")
//	t.textCenter(5, "от")
//	t.form(32.5, "")
//	t.paragraph("(5б)")
//
//	t.document.Ln(2 * lineHeight)
//	t.SavePosition()
//	t.LineHeaderV()
//
//	t.document.SetLeftMargin(margin + separatorLeftMargin + labelWidth + formWidth + paragraphWidth + 10)
//	t.document.SetY(2*lineHeight + 5 + topMargin)
//
//	t.labelBoldFormMultiline(labelWidth, "Покупатель:", formWidth, "", "(6)")
//	t.labelFormMultiline(labelWidth, "Адрес:", formWidth, "", "(6а)")
//	t.labelForm(labelWidth, "ИНН/КПП покупателя:", formWidth, "", "(6б)")
//	t.labelForm(labelWidth, "Валюта: наименование, код:", formWidth, "", "(7)")
//	t.text(labelWidth+10, "Идентификатор государственного контракта,")
//	t.nl()
//	t.text(labelWidth+10, "договора (соглашения) (при наличии):")
//	t.form(formWidth-10, "")
//	t.paragraph("(8)")
//}
//
//func (t *Template) LineHeaderV() {
//	t.document.SetLineWidth(0.3)
//	t.document.Line(margin+separatorLeftMargin+2, topMargin, margin+separatorLeftMargin+2, t.posY)
//	t.document.SetLineWidth(0.1)
//}
//
//func (t *Template) LineTableV(y, h float64) {
//	t.document.SetLineWidth(0.3)
//	t.document.Line(margin+separatorLeftMargin+2, y, margin+separatorLeftMargin+2, y+h)
//	t.document.SetLineWidth(0.1)
//}
//
//func (t *Template) TableHeader() {
//	t.document.SetLeftMargin(margin)
//	t.document.SetY(t.posY)
//
//	const (
//		headerCellHeight         = 26
//		headerMergedCellHeight   = 10
//		headerSeparateCellHeight = headerCellHeight - headerMergedCellHeight
//	)
//
//	t.LineTableV(t.posY, headerCellHeight)
//	t.tableHeaderCell(t.table.widths["А"], headerCellHeight, "Код\nтовара/работ,\nуслуг")
//	t.tableHeaderCell(t.table.widths["1"], headerCellHeight, "№\nп/п")
//	t.tableHeaderCell(t.table.widths["1а"], headerCellHeight, "Наименование товара\n(описание выполненных работ, оказанных услуг),\nимущественного права")
//	t.tableHeaderCell(t.table.widths["1б"], headerCellHeight, "Код\nвида\nтовара")
//	t.tableHeaderCell(t.table.widths["2"]+t.table.widths["2а"], headerMergedCellHeight, "Единица\nизмерения")
//	t.tableHeaderCell(t.table.widths["3"], headerCellHeight, "Количество\n(объем)")
//	t.tableHeaderCell(t.table.widths["4"], headerCellHeight, "Цена\n(тариф) за\nединицу\nизмерения")
//	t.tableHeaderCell(t.table.widths["5"], headerCellHeight, "Стоимость\nтоваров (работ,\nуслуг),\nимущественных\nправ без налога -\nвсего")
//	t.tableHeaderCell(t.table.widths["6"], headerCellHeight, "В том числе\nсумма\nакциза")
//	t.tableHeaderCell(t.table.widths["7"], headerCellHeight, "Налоговая\nставка")
//	t.tableHeaderCell(t.table.widths["8"], headerCellHeight, "Сумма налога,\nпредъявляемая\nпокупателю")
//	t.tableHeaderCell(t.table.widths["9"], headerCellHeight, "Стоимость\nтоваров (работ,\nуслуг),\nимущественных\nправ с налогом -\nвсего")
//	t.tableHeaderCell(t.table.widths["10"]+t.table.widths["10а"], headerMergedCellHeight, "Страна\nпроисхождения\nтовара")
//	t.tableHeaderCell(t.table.widths["11"], headerCellHeight, "Регистрационный\nномер декларации\nна товары или\nрегистрационный\nномер партии\nтовара,\nподлежащего\nпрослеживаемости")
//
//	t.document.SetLeftMargin(margin + t.table.widths["А"] + t.table.widths["1"] + t.table.widths["1а"] + t.table.widths["1б"])
//	t.document.SetY(t.posY + headerMergedCellHeight)
//
//	t.tableHeaderCell(t.table.widths["2"], headerSeparateCellHeight, "код")
//	t.tableHeaderCell(t.table.widths["2а"], headerSeparateCellHeight, "условное\nобозна-\nчение\n(нацио-\nнальное)")
//
//	t.document.SetLeftMargin(a4LandscapeWidth - (margin + t.table.widths["11"] + t.table.widths["10а"] + t.table.widths["10"]))
//
//	t.tableHeaderCell(t.table.widths["10"], headerSeparateCellHeight, "цифровой\nкод")
//	t.tableHeaderCell(t.table.widths["10а"], headerSeparateCellHeight, "краткое\nнаимено-\nвание")
//
//	t.document.SetY(t.posY + headerCellHeight)
//	t.SavePosition()
//}
//
//func (t *Template) TableLetterHeader() {
//	if !t.table.showHeader {
//		return
//	}
//
//	t.document.SetX(margin)
//
//	const headerCellHeight = 3
//
//	t.LineTableV(t.document.GetY(), headerCellHeight)
//	t.tableHeaderLetterCell(headerCellHeight, "А")
//	t.tableHeaderLetterCell(headerCellHeight, "1")
//	t.tableHeaderLetterCell(headerCellHeight, "1а")
//	t.tableHeaderLetterCell(headerCellHeight, "1б")
//	t.tableHeaderLetterCell(headerCellHeight, "2")
//	t.tableHeaderLetterCell(headerCellHeight, "2а")
//	t.tableHeaderLetterCell(headerCellHeight, "3")
//	t.tableHeaderLetterCell(headerCellHeight, "4")
//	t.tableHeaderLetterCell(headerCellHeight, "5")
//	t.tableHeaderLetterCell(headerCellHeight, "6")
//	t.tableHeaderLetterCell(headerCellHeight, "7")
//	t.tableHeaderLetterCell(headerCellHeight, "8")
//	t.tableHeaderLetterCell(headerCellHeight, "9")
//	t.tableHeaderLetterCell(headerCellHeight, "10")
//	t.tableHeaderLetterCell(headerCellHeight, "10а")
//	t.tableHeaderLetterCell(headerCellHeight, "11")
//
//	t.document.Ln(headerCellHeight)
//	t.SavePosition()
//}
//
//func (t *Template) TableBodyCells(item HTMLDetail) {
//	t.document.SetXY(margin, t.posY)
//
//	titleLines := t.document.SplitText(item.Title, t.table.widths["1а"])
//	t.table.currentRowHeight = float64(len(titleLines) * lineHeight)
//
//	t.tableBodyCell(t.table.widths["А"], t.table.currentRowHeight, item.Code, "BC")
//	t.tableBodyCell(t.table.widths["1"], t.table.currentRowHeight, item.Number, "BC")
//	t.tableBodyMultiLineCell(t.table.widths["1а"], lineHeight, item.Title, "BL")
//	t.tableBodyCell(t.table.widths["1б"], t.table.currentRowHeight, item.KindID, "BR")
//	t.tableBodyCell(t.table.widths["2"], t.table.currentRowHeight, item.OkeiID, "BR")
//	t.tableBodyCell(t.table.widths["2а"], t.table.currentRowHeight, item.OkeiCode, "BL")
//	t.tableBodyCell(t.table.widths["3"], t.table.currentRowHeight, item.Quantity, "BR")
//	t.tableBodyCell(t.table.widths["4"], t.table.currentRowHeight, item.Price, "BR")
//	t.tableBodyCell(t.table.widths["5"], t.table.currentRowHeight, item.AmountWithoutVat, "BR")
//	t.tableBodyCell(t.table.widths["6"], t.table.currentRowHeight, item.Excise, "BR")
//	t.tableBodyCell(t.table.widths["7"], t.table.currentRowHeight, item.Vat, "BR")
//	t.tableBodyCell(t.table.widths["8"], t.table.currentRowHeight, item.AmountVat, "BR")
//	t.tableBodyCell(t.table.widths["9"], t.table.currentRowHeight, item.AmountWithVat, "BR")
//	t.tableBodyCell(t.table.widths["10"], t.table.currentRowHeight, item.CountryID, "BR")
//	t.tableBodyCell(t.table.widths["10а"], t.table.currentRowHeight, item.CountryName, "BL")
//	t.tableBodyCell(t.table.widths["11"], t.table.currentRowHeight, item.Gtd, "BR")
//
//	t.nl()
//	t.LineTableV(t.posY, t.table.currentRowHeight)
//	t.table.ResetRowHeight()
//	t.SavePosition()
//}
//
//func (t *Template) TableFooterCells() {
//	t.document.SetXY(margin, t.posY)
//
//	t.document.SetFontStyle("B")
//	defer t.document.SetFontStyle("")
//
//	t.tableBodyCell(t.table.widths["А"], lineHeight, "", "")
//	t.tableBodyCell(t.table.totalToPayWidth(), lineHeight, "Всего к оплате", "BL")
//	t.tableBodyCell(t.table.widths["5"], lineHeight, t.entity.AmountWithoutVatTotal, "BR")
//	t.tableBodyCell(t.table.totalXWidth(), lineHeight, "X", "BC")
//	t.tableBodyCell(t.table.widths["8"], lineHeight, t.entity.AmountVatTotal, "BR")
//	t.tableBodyCell(t.table.widths["9"], lineHeight, t.entity.AmountWithVatTotal, "BR")
//	t.tableBodyCell(t.table.totalEndWidth(), lineHeight, "", "")
//
//	t.LineTableV(t.posY, lineHeight)
//	t.document.Ln(2 * lineHeight)
//	t.table.showHeader = false
//	t.SavePosition()
//}
//
//func (t *Template) Signatures() {
//	t.document.SetLeftMargin(margin + separatorLeftMargin + 5)
//	t.document.SetY(t.posY)
//
//	const (
//		labelW = 46
//		formW  = 42
//		gapW   = 2
//	)
//
//	t.textMultiLine(labelW, "Руководитель организации\nили иное уполномоченное лицо")
//	t.formCH(formW, 2*lineHeight, "")
//	t.gap(gapW)
//	t.formCH(formW, 2*lineHeight, t.entity.OrgChiefName)
//	t.gap(gapW)
//	t.textMultiLine(labelW, "Главный бухгалтер\nили иное уполномоченное лицо")
//	t.formCH(formW, 2*lineHeight, "")
//	t.gap(gapW)
//	t.formCH(formW, 2*lineHeight, t.entity.OrgAccountantName)
//
//	t.nl()
//	t.document.SetFontSize(subFontSize)
//	t.gap(labelW)
//	t.textUnderscore(formW, "(подпись)")
//	t.gap(gapW)
//	t.textUnderscore(formW, "(ф.и.о.)")
//	t.gap(gapW + labelW)
//	t.textUnderscore(formW, "(подпись)")
//	t.gap(gapW)
//	t.textUnderscore(formW, "(ф.и.о.)")
//	t.document.SetFontSize(mainFontSize)
//
//	t.nl()
//	t.textMultiLine(labelW, "Индивидуальный предприниматель\nили иное уполномоченное лицо")
//	t.formCH(formW, 2*lineHeight, "")
//	t.gap(gapW)
//	t.formCH(formW, 2*lineHeight, "")
//	t.gap(gapW)
//	t.formCH(formW*2+labelW+gapW, 2*lineHeight, "")
//
//	t.nl()
//	t.document.SetFontSize(subFontSize)
//	t.gap(labelW)
//	t.textUnderscore(formW, "(подпись)")
//	t.gap(gapW)
//	t.textUnderscore(formW, "(ф.и.о.)")
//	t.gap(gapW)
//	t.textUnderscore(formW*gapW+labelW+2, "(основной государственный регистрационный номер индивидуального предпринимателя и дата присвоения такого номера)")
//	t.document.SetFontSize(mainFontSize)
//
//	t.document.Ln(lineHeight)
//	h := t.document.GetY() - t.posY
//	t.document.SetY(t.posY)
//	t.document.Cell(0, h, strconv.FormatFloat(h, 'f', 0, 64))
//	t.nl()
//	t.LineSignatures()
//	t.document.Ln(lineHeight)
//	t.SavePosition()
//}
//
//func (t *Template) LineSignatures() {
//	y := t.document.GetY()
//
//	t.document.SetLineWidth(0.3)
//	t.document.Line(margin+separatorLeftMargin+2, t.posY-lineHeight, margin+separatorLeftMargin+2, y)
//	t.document.Line(margin+separatorLeftMargin+2, y, a4LandscapeWidth-margin, y)
//	t.document.SetLineWidth(0.1)
//}
//
//func (t *Template) gap(w float64) {
//	t.document.Cell(w, lineHeight, "")
//}
//
//func (t *Template) paragraph(text string) {
//	t.document.CellFormat(paragraphWidth, lineHeight, text, "0", 1, "BC", false, 0, "")
//}
//
//func (t *Template) cellFormat(w, h float64, text, borderStr string, ln int, align string) {
//	t.document.CellFormat(w, h, text, borderStr, ln, align, false, 0, "")
//}
//
//func (t *Template) form(w float64, text string) {
//	t.document.CellFormat(w, lineHeight, text, "B", 0, "BL", false, 0, "")
//}
//
//func (t *Template) formCH(w, h float64, text string) {
//	t.document.CellFormat(w, h, text, "B", 0, "BC", false, 0, "")
//}
//
//func (t *Template) formCenter(w float64, text string) {
//	t.document.CellFormat(w, lineHeight, text, "B", 0, "BC", false, 0, "")
//}
//
//func (t *Template) formMultiLine(w float64, text string) {
//	x, y := t.document.GetXY()
//	t.document.MultiCell(w, lineHeight, text, "B", "BL", false)
//	t.document.SetXY(x+w, y)
//}
//
//func (t *Template) text(w float64, text string) {
//	t.document.CellFormat(w, lineHeight, text, "0", 0, "BL", false, 0, "")
//}
//
//func (t *Template) textBold(w float64, text string) {
//	t.document.SetFontStyle("B")
//	t.document.CellFormat(w, lineHeight, text, "0", 0, "BL", false, 0, "")
//	t.document.SetFontStyle("")
//}
//
//func (t *Template) textCenter(w float64, text string) {
//	t.document.CellFormat(w, lineHeight, text, "0", 0, "BC", false, 0, "")
//}
//
//func (t *Template) textMultiLine(w float64, text string) {
//	x, y := t.document.GetXY()
//	t.document.MultiCell(w, lineHeight, text, "0", "BL", false)
//	t.document.SetXY(x+w, y)
//}
//
//func (t *Template) textUnderscore(w float64, text string) {
//	t.document.CellFormat(w, subLineHeight, text, "0", 0, "TC", false, 0, "")
//}
//
//func (t *Template) tableHeaderCell(w, h float64, text string) {
//	textParts := t.document.SplitText(text, w)
//
//	x, y := t.document.GetXY()
//
//	t.document.Rect(x, y, w, h, "D")
//
//	textY := y + (h-float64(len(textParts)*lineHeight))/2
//
//	t.document.SetY(textY)
//
//	for _, textPart := range textParts {
//		t.textCenter(w, textPart)
//		t.nl()
//	}
//
//	t.document.SetLeftMargin(x + w)
//	t.document.SetY(y)
//}
//
//func (t *Template) tableHeaderLetterCell(h float64, text string) {
//	t.cellFormat(t.table.widths[text], h, text, "1", 0, "CM")
//}
//
//func (t *Template) tableBodyCell(w, h float64, text string, align string) {
//	t.cellFormat(w, h, text, "1", 0, align)
//}
//
//func (t *Template) tableBodyMultiLineCell(w, h float64, text string, align string) {
//	x, y := t.document.GetXY()
//
//	t.document.MultiCell(w, h, text, "1", align, false)
//
//	t.table.currentRowHeight = t.document.GetY() - y
//
//	t.document.SetXY(x+w, y)
//}
//
//func (t *Template) labelForm(wl float64, label string, wf float64, text, paragraph string) {
//	t.text(wl, label)
//	t.form(wf, text)
//	t.paragraph(paragraph)
//}
//
//func (t *Template) labelFormMultiline(wl float64, label string, wf float64, text, paragraph string) {
//	t.text(wl, label)
//	t.formMultiLine(wf, text)
//	t.paragraph(paragraph)
//}
//
//func (t *Template) labelBoldForm(wl float64, label string, wf float64, text, paragraph string) {
//	t.textBold(wl, label)
//	t.form(wf, text)
//	t.paragraph(paragraph)
//}
//
//func (t *Template) labelBoldFormMultiline(wl float64, label string, wf float64, text, paragraph string) {
//	t.textBold(wl, label)
//	t.formMultiLine(wf, text)
//	t.paragraph(paragraph)
//}
//
//func (t *Template) nl() {
//	t.document.Ln(-1)
//}
//
//func (t *table) totalToPayWidth() (width float64) {
//	width = width +
//		+t.widths["1"] +
//		+t.widths["1а"] +
//		+t.widths["1б"] +
//		+t.widths["2"] +
//		+t.widths["2а"] +
//		+t.widths["3"] +
//		+t.widths["4"]
//
//	return width
//}
//
//func (t *table) totalXWidth() (width float64) {
//	width = width +
//		+t.widths["6"] +
//		+t.widths["7"]
//
//	return width
//}
//
//func (t *table) totalEndWidth() (width float64) {
//	width = width +
//		+t.widths["10"] +
//		+t.widths["10а"] +
//		+t.widths["11"]
//
//	return width
//}
//
//func getTableWidths() map[string]float64 {
//	tableWidths := map[string]float64{
//		"А": separatorLeftMargin + 2,
//		"1": 8,
//		//"1а":  50,
//		"1б":  10,
//		"2":   10,
//		"2а":  14,
//		"3":   15,
//		"4":   20,
//		"5":   22,
//		"6":   15,
//		"7":   12,
//		"8":   22,
//		"9":   22,
//		"10":  14,
//		"10а": 13,
//		"11":  24,
//	}
//
//	totalTableWidth := 0.0
//
//	for _, width := range tableWidths {
//		totalTableWidth += width
//	}
//
//	tableWidths["1а"] = a4LandscapeWidth - totalTableWidth - margin*2
//
//	return tableWidths
//}
