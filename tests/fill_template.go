package tests

import (
	"fmt"
	"strconv"

	"github.com/eugene-static/pdfego"
	"github.com/eugene-static/pdfego/unit"
)

func initCore() (*pdfego.Core, error) {
	c := pdfego.NewCore(pdfego.Landscape)
	c.SetMargins(3, 3, 3, 8)
	c.SetDefaultFontSize(6)
	c.SetDefaultBorderSize(0.3)
	c.EnableCompression()
	c.IgnoreImageNotFound()

	err := c.SetFontRegular("./fonts/LiberationSans-Regular.ttf")
	if err != nil {
		return nil, err
	}

	err = c.SetFontBold("./fonts/LiberationSans-Bold.ttf")
	if err != nil {
		return nil, err
	}

	err = c.SetFontItalic("./fonts/LiberationSans-Italic.ttf")
	if err != nil {
		return nil, err
	}

	err = c.ReadImage("fisher.png", "fisher")
	if err != nil {
		return nil, err
	}

	err = c.ReadImage("newell.png", "gaben")
	if err != nil {
		return nil, err
	}

	return c, nil
}

type updTemplate struct {
	core         *pdfego.Core
	view         View
	tableColumns []unit.MM
}

func newUpdTemplate(core *pdfego.Core, view View) *updTemplate {
	return &updTemplate{
		core:         core,
		view:         view,
		tableColumns: []unit.MM{21, 7, 83, 7, 7, 10, 15, 15, 20, 13, 13, 20, 20, 8, 10, 22},
	}
}

func (tmpl *updTemplate) fill() ([]byte, error) {
	iterator := pdfego.NewIterator(tmpl.view.Details, tmpl.tableDetailsIter)

	constructor := pdfego.NewConstructor(tmpl.core)

	constructor.
		Paginator().
		Apply(pdfego.DefaultPaginator)

	constructor.
		Watermark(pdfego.WatermarkOptions{Align: "RB"}).
		Apply(tmpl.watermark)

	constructor.
		Build(tmpl.header).
		Build(tmpl.tableHeader).
		ApplyHeader(tmpl.tableNumberHeader).
		Iterate(iterator).
		ReleaseHeader().
		Build(tmpl.tableFooter).
		Build(tmpl.signatories).
		Build(tmpl.shippingHeaders).
		Build(tmpl.shippingBody)

	return constructor.Bytes()
}

func (tmpl *updTemplate) header(block *pdfego.Block) {
	table := block.Slot().Table(3, pdfego.Columns(15, 5))

	table.Row().
		Cell("Универсальный передаточный документ", pdfego.CellOptions{Height: 10, Wrap: true, Colspan: 2, Align: "LT"})
	table.Row().
		Cell("Статус", pdfego.CellOptions{Height: 5, Align: "LM"}).
		Cell(tmpl.view.Status, pdfego.CellOptions{Height: 5, Align: "CM", Border: "O"})
	table.Row().
		Cell("1 - счет-фактура и\nпередаточный\nдокумент (акт)\n2 - передаточный\nдокумент (акт)", pdfego.CellOptions{Height: 15, Align: "LB", Colspan: 2, FontSize: 5, Wrap: true})

	numberDatesRequisitesSlot := block.Slot(
		pdfego.NodeOptions{
			IndentH: 1,
			Ledge:   5,
			Border:  "L",
		},
	)

	numbersDates := numberDatesRequisitesSlot.Block(
		pdfego.NodeOptions{
			IndentH: 1,
		},
	)

	table = numbersDates.Slot().
		Table(2, []unit.MM{15, 5, 20, 5, 20, 10})

	table.Row().
		Label("Счет-фактура").
		Cell("№", pdfego.CellOptions{Align: "CB"}).
		Blank(tmpl.view.SfNum, "").
		Cell("от", pdfego.CellOptions{Align: "CB"}).
		Blank(tmpl.view.SfDate, "").
		Paragraph("(1)")

	table.Row().
		Label("Исправление").
		Cell("№", pdfego.CellOptions{Align: "CB"}).
		BlankEmpty().
		Cell("от", pdfego.CellOptions{Align: "CB"}).
		BlankEmpty().
		Paragraph("(1а)")

	table = numbersDates.
		Slot().
		Table(1, []unit.MM{190})

	table.Row().
		Cell("Приложение № 1 к постановлению Правительства Российской Федерации от 26 декабря 2011 г. № 1137\n(в редакции постановления Правительства Российской Федерации от 16 августа 2024 г. № 1096)",
			pdfego.CellOptions{Align: "RT", Font: pdfego.FontItalic, FontSize: 5, Wrap: true})

	requisites := numberDatesRequisitesSlot.
		Block(pdfego.NodeOptions{
			IndentH: 1,
			IndentV: 5,
		},
		)

	requisitesColumns := []unit.MM{45, 80, 10}

	table = requisites.
		Slot().
		Table(10, requisitesColumns)

	table.Row().
		LabelHead("Продавец:").
		Form(tmpl.view.OrgPrintName, "LB").
		Paragraph("(2)")
	table.Row().
		Label("Адрес:").
		Form(tmpl.view.OrgPrintAddress, "LB").
		Paragraph("(2а)")
	table.Row().
		Label("ИНН/КПП продавца:").
		Blank(tmpl.view.OrgInnKpp, "LB").
		Paragraph("(2б)")
	table.Row().
		Label("Грузоотправитель и его адрес:").
		Form(tmpl.view.ShipperPrintNameAddress, "LB").
		Paragraph("(3)")
	table.Row().
		Label("Грузополучатель и его адрес:").
		Form(tmpl.view.ConsigneePrintNameAddress, "LB").
		Paragraph("(4)")
	table.Row().
		Label("К платежно-расчетному документу №:").
		Blank(tmpl.view.PaymentAndSettlementDocument, "LB").
		Paragraph("(5)")
	table.Row().
		Label("Документ об отгрузке:").
		Form(tmpl.view.ShippingDocuments, "LB").
		Paragraph("(5а)")
	table.Row().
		LabelSpan(
			"К счету-фактуре (счетам-фактурам), выставленному (выставленным)\n"+
				"при получении оплаты, частичной оплаты или иных платежей в счет\n"+
				"предстоящих поставок товаров (выполнения работ, оказания услуг),",
			2,
			pdfego.CellOptions{Wrap: true},
		)
	table.Row().
		Label("передачи имущественных прав №:").
		BlankEmpty()
	table.Row().
		Label("исправление №:").
		BlankEmpty().
		Paragraph("(5б)")

	table = requisites.Slot().
		Table(5, requisitesColumns)

	table.Row().
		LabelHead("Покупатель:").
		Form(tmpl.view.SupplierPrintName, "LB").
		Paragraph("(6)")
	table.Row().
		Label("Адрес:").
		Form(tmpl.view.SupplierPrintAddress, "LB").
		Paragraph("(6а)")
	table.Row().
		Label("ИНН/КПП покупателя:").
		Blank(tmpl.view.SupplierInnKpp, "LB").
		Paragraph("(6б)")
	table.Row().
		Label("Валюта: наименование, код:").
		Blank(tmpl.view.CurrencyNameCode, "LB").
		Paragraph("(7)")
	table.Row().
		Label("Идентификатор государственного контракта,\nдоговора (соглашения) (при наличии):", pdfego.CellOptions{Wrap: true}).
		BlankEmpty().
		Paragraph("(8)")
}

func (tmpl *updTemplate) tableHeader(block *pdfego.Block) {
	table := block.Slot().Table(2, tmpl.tableColumns)

	optsBounded := pdfego.CellOptions{
		Height: 15,
		Border: "o",
		Wrap:   true,
	}

	optsBoundedRS2 := pdfego.CellOptions{
		Rowspan: 2,
		Border:  "o",
		Wrap:    true,
	}

	optsBoundedCS2 := pdfego.CellOptions{
		Colspan: 2,
		Border:  "o",
		Wrap:    true,
	}

	table.Row().
		Cell("Код\nтовара/работ, услуг", pdfego.CellOptions{Height: 10, Border: "tblR", Wrap: true, Rowspan: 2}).
		Cell("№\nп/п", optsBoundedRS2).
		Cell("Наименование товара\n(описание выполненных работ, оказанных услуг),\nимущественного права", optsBoundedRS2).
		Cell("Код\nвида\nтовара", optsBoundedRS2).
		Cell("Единица\nизмерения", optsBoundedCS2).
		Cell("Количество\n(объем)", optsBoundedRS2).
		Cell("Цена\n(тариф) за\nединицу\nизмерения", optsBoundedRS2).
		Cell("Стоимость\nтоваров (работ,\nуслуг),\nимущественных\nправ без налога -\nвсего", optsBoundedRS2).
		Cell("В том числе сумма акциза", optsBoundedRS2).
		Cell("Налоговая\nставка", optsBoundedRS2).
		Cell("Сумма налога,\nпредъявляемая\nпокупателю", optsBoundedRS2).
		Cell("Стоимость\nтоваров (работ,\nуслуг),\nимущественных\nправ с налогом -\nвсего", optsBoundedRS2).
		Cell("Страна\nпроисхождения\nтовара", optsBoundedCS2).
		Cell("Регистрационный\nномер декларации\nна товары или\nрегистрационный\nномер партии\nтовара,\nподлежащего\nпрослеживаемости", optsBoundedRS2)

	table.Row().
		Cell("код", optsBounded).
		Cell("условное обозна-\nчение\n(нацио-\nнальное)", optsBounded).
		Cell("цифро-\nвой\nкод", optsBounded).
		Cell("краткое\nнаимено-\nвание", optsBounded)
}

func (tmpl *updTemplate) tableNumberHeader(block *pdfego.Block) {
	table := block.Slot().
		Table(1, tmpl.tableColumns)

	table.Row().
		Cell("А", pdfego.CellOptions{Align: "CM", Height: 3, Border: "tblR"}).
		Outlined("1").
		Outlined("1а").
		Outlined("1б").
		Outlined("2").
		Outlined("2а").
		Outlined("3").
		Outlined("4").
		Outlined("5").
		Outlined("6").
		Outlined("7").
		Outlined("8").
		Outlined("9").
		Outlined("10").
		Outlined("10а").
		Outlined("11")
}

func (tmpl *updTemplate) tableDetailsIter(block *pdfego.Block, item Detail, index int) {
	table := block.Slot().
		Table(1, tmpl.tableColumns)

	table.Row().
		Cell(item.Code, pdfego.CellOptions{Align: "CB", Border: "tblR"}).
		Outlined(strconv.Itoa(index+1), pdfego.CellOptions{Align: "CB"}).
		Outlined(item.Title, pdfego.CellOptions{Align: "LB", Wrap: true}).
		Outlined(item.KindID, pdfego.CellOptions{Align: "CB"}).
		Outlined(item.OkeiID, pdfego.CellOptions{Align: "RB"}).
		Outlined(item.OkeiCode, pdfego.CellOptions{Align: "LB"}).
		Outlined(item.Quantity, pdfego.CellOptions{Align: "RB"}).
		Outlined(item.Price, pdfego.CellOptions{Align: "RB"}).
		Outlined(item.AmountWithoutVat, pdfego.CellOptions{Align: "RB"}).
		Outlined(item.Excise, pdfego.CellOptions{Align: "RB"}).
		Outlined(item.Vat, pdfego.CellOptions{Align: "RB"}).
		Outlined(item.AmountVat, pdfego.CellOptions{Align: "RB"}).
		Outlined(item.AmountWithVat, pdfego.CellOptions{Align: "RB"}).
		Outlined(item.CountryID, pdfego.CellOptions{Align: "RB"}).
		Outlined(item.CountryName, pdfego.CellOptions{Align: "LB"}).
		Outlined(item.Gtd, pdfego.CellOptions{Align: "LB"})
}

func (tmpl *updTemplate) tableFooter(block *pdfego.Block) {
	table := block.Slot().
		Table(1, tmpl.tableColumns)

	table.Row().
		Cell("", pdfego.CellOptions{Border: "tblR", Height: 3}).
		Outlined("Всего к оплате:", pdfego.CellOptions{Align: "RB", Colspan: 7, Font: pdfego.FontBold}).
		Outlined(tmpl.view.AmountWithoutVatTotal, pdfego.CellOptions{Align: "RB"}).
		Outlined("X", pdfego.CellOptions{Align: "CB", Colspan: 2, Font: pdfego.FontBold}).
		Outlined(tmpl.view.AmountVatTotal, pdfego.CellOptions{Align: "RB"}).
		Outlined(tmpl.view.AmountWithVatTotal, pdfego.CellOptions{Align: "RB"}).
		Outlined("", pdfego.CellOptions{Colspan: 3})
}

func (tmpl *updTemplate) signatories(block *pdfego.Block) {
	table := block.
		Slot().
		Table(2, pdfego.Columns(3, 6, 9), pdfego.TableOptions{SpacingH: 1})

	table.Row().
		Cell("Документ составлен", pdfego.CellOptions{Colspan: 3, Align: "LB"})
	table.Row().
		Cell("на", pdfego.CellOptions{Align: "LB"}).
		PagesCount(pdfego.CellOptions{Align: "CB", Border: "b"}).
		Cell("листах", pdfego.CellOptions{Align: "LB"})

	table = block.Slot(pdfego.NodeOptions{
		IndentH: 1,
		Border:  "LB",
		Ledge:   3,
	}).
		Table(
			4,
			[]unit.MM{38, 47, 47, 38, 47, 47},
			pdfego.TableOptions{SpacingH: 1, IndentH: 1})

	table.Row().
		Cell("Руководитель организации\nили иное уполномоченное лицо", pdfego.CellOptions{Height: 10, Align: "LB", Wrap: true}).
		Blank("[электронная подпись]", "CB").
		Blank(tmpl.view.OrgChiefName, "CB").
		Cell("Главный бухгалтер\nили иное уполномоченное лицо", pdfego.CellOptions{Height: 10, Align: "LB", Wrap: true}).
		Blank("[электронная подпись]", "CB").
		Blank(tmpl.view.OrgAccountantName, "CB")
	table.Row().
		Skip().
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)").
		Skip().
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)")
	table.Row().
		Cell("Индивидуальный предприниматель\nили иное уполномоченное лицо", pdfego.CellOptions{Height: 10, Align: "LB", Wrap: true}).
		BlankEmpty().
		BlankEmpty().
		BlankEmpty(pdfego.CellOptions{Colspan: 3})
	table.Row().
		Skip().
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)").
		UnderscoreSpan("(основной государственный регистрационный номер индивидуального предпринимателя и дата присвоения такого номера)", 3)
}

func (tmpl *updTemplate) shippingHeaders(block *pdfego.Block) {
	table := block.Slot().
		Table(4, []unit.MM{60, 224, 10},
			pdfego.TableOptions{
				IndentV: 5,
			},
		)

	table.Row().
		Label("Основание передачи (сдачи) / получения (приемки)").
		Blank(tmpl.view.Contract, "LB").
		Paragraph("[10]")
	table.Row().
		Skip().
		Underscore("(договор; доверенность и др.)")
	table.Row().
		Label("Данные о транспортировке и грузе").
		BlankEmpty().
		Paragraph("[11]")
	table.Row().
		Skip().
		Underscore("(транспортная накладная, поручение экспедитору, экспедиторская / складская расписка и др. / масса нетто/ брутто груза, если не приведены ссылки на транспортные документы, содержащие эти сведения)")
}

func (tmpl *updTemplate) shippingBody(block *pdfego.Block) {
	columns := []unit.MM{44, 44, 44, 10}
	rowOptions := pdfego.RowOptions{MinHeight: 4}

	table := block.Slot(
		pdfego.NodeOptions{
			IndentV: 3,
		},
	).
		Table(14, columns,
			pdfego.TableOptions{
				SpacingH: 1,
			},
		)

	table.Row().
		LabelSpan("Товар (груз) передал / услуги, результаты работ, права сдал", 2)
	table.Row(rowOptions).
		Blank(tmpl.view.StoreKeeperPosition, "").
		Blank("[электронная подпись]", "").
		Blank(tmpl.view.StoreKeeperName, "").
		Paragraph("[12]")
	table.Row().
		Underscore("(должность)").
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)")
	table.Row(rowOptions).
		Label("Дата отгрузки, передачи (сдачи)").
		BlankSpan(tmpl.view.DocSendDate, "", 2).
		Paragraph("[13]")
	table.Row(rowOptions).
		LabelSpan("Иные сведения об отгрузке, передаче", 2)
	table.Row(rowOptions).
		BlankSpan(tmpl.view.LkID, "LB", 3).
		Paragraph("[14]")
	table.Row().
		UnderscoreSpan("(ссылки на неотъемлемые приложения, сопутствующие документы, иные документы и т.п.)", 3)
	table.Row(rowOptions).
		LabelSpan("Ответственный за правильность оформления факта хозяйственной жизни", 2)
	table.Row(rowOptions).
		Blank(tmpl.view.SenderChiefPosition, "").
		Blank("[электронная подпись]", "").
		Blank(tmpl.view.SenderChiefName, "").
		Paragraph("[15]")
	table.Row().
		Underscore("(должность)").
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)")
	table.Row(rowOptions).
		LabelSpan("Наименование экономического субъекта – составителя документа (в т.ч. комиссионера / агента)", 2)
	table.Row(rowOptions).
		BlankSpan(tmpl.view.OrgPrintName, "LB", 3).
		Paragraph("[16]")
	table.Row().
		UnderscoreSpan("(может не заполняться при проставлении печати в М.П., может быть указан ИНН / КПП)", 3)
	table.Row().
		Cell("М.П.")

	table = block.Slot(
		pdfego.NodeOptions{
			IndentV: 3,
			IndentH: 1,
			Border:  "L",
		},
	).
		Table(14, columns,
			pdfego.TableOptions{
				SpacingH: 1,
				IndentH:  3,
			},
		)

	table.Row().
		LabelSpan("Товар (груз) получил / услуги, результаты работ, права принял", 2)
	table.Row(rowOptions).
		Blank(tmpl.view.RecipientPosition, "").
		Image("fisher", pdfego.CellOptions{Border: "b", Align: "BC", Scale: 2, OffsetV: 2, PlaceHolder: "Jeliy Fisher"}).
		Blank(tmpl.view.RecipientName, "").
		Paragraph("[17]")
	table.Row().
		Underscore("(должность)").
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)")
	table.Row(rowOptions).
		Label("Дата получения (приемки)").
		BlankSpan(tmpl.view.DocReceiveDate, "", 2).
		Paragraph("[18]")
	table.Row(rowOptions).
		LabelSpan("Иные сведения о получении, приемке", 2)
	table.Row(rowOptions).
		BlankSpan("", "LB", 3).
		Paragraph("[19]")
	table.Row().
		UnderscoreSpan("(информация о наличии/отсутствии претензии; ссылки на неотъемлемые приложения, и другие документы и т.п.)\nОтветственный", 3)
	table.Row(rowOptions).
		LabelSpan("Ответственный за правильность оформления факта хозяйственной жизни", 2)
	table.Row(rowOptions).
		Blank(tmpl.view.RecipientChiefPosition, "").
		Image("gaben", pdfego.CellOptions{Border: "b", Align: "BC", Scale: 2, OffsetV: 2, PlaceHolder: "Gabe Newell"}).
		Blank(tmpl.view.RecipientChiefName, "").
		Paragraph("[20]")
	table.Row().
		Underscore("(должность)").
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)")
	table.Row(rowOptions).
		LabelSpan("Наименование экономического субъекта – составителя документа (в т.ч. комиссионера / агента)", 2)
	table.Row(rowOptions).
		BlankSpan(tmpl.view.SupplierPrintName, "LB", 3).
		Paragraph("[21]")
	table.Row().
		UnderscoreSpan("(может не заполняться при проставлении печати в М.П., может быть указан ИНН / КПП)", 3)
	table.Row().
		Cell("М.П.")
}

func (tmpl *updTemplate) watermark(block *pdfego.Block) {
	table := block.Slot().
		Table(3, []unit.MM{40, 40, 40, 40},
			pdfego.TableOptions{
				SpacingH:    1,
				SpacingV:    1,
				Border:      "O",
				TextColor:   pdfego.ColorDarkBlue,
				BorderColor: pdfego.ColorDarkBlue,
			})

	opts := pdfego.CellOptions{
		FontSize: 5,
	}

	table.Row().
		Skip().
		Label("Тип подписи, сотрудник", opts).
		Label("Серийный номер, период действия", opts).
		Label("Дата и время подписи", opts)
	table.Row().
		Label("Подпись отправителя", opts).
		Label("Квалифицированная ЭП", opts).
		Label("номер КЭП"+tmpl.view.Certificate.KEP, opts).
		Label(tmpl.view.Certificate.SignDate, opts)
	table.Row().
		Skip().
		Label(tmpl.view.Certificate.SignerName, opts).
		Label(fmt.Sprintf("период действия с %s\nпо %s", tmpl.view.Certificate.ValidityPeriodFrom, tmpl.view.Certificate.ValidityPeriodTo), pdfego.CellOptions{FontSize: 5, Wrap: true})
}
