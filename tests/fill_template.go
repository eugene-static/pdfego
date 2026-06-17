package tests

import (
	"fmt"

	"gitlab.wildberries.ru/ovp/ovp/go-infrastructure/pdfego"
	"gitlab.wildberries.ru/ovp/ovp/go-infrastructure/pdfego/unit"
)

func prepareTemplate() (*updTemplate, error) {
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

	err = c.ReadImage("fisher.png", "fisher")
	if err != nil {
		return nil, err
	}

	err = c.ReadImage("newell.png", "gaben")
	if err != nil {
		return nil, err
	}

	return &updTemplate{
		constructor:  pdfego.NewConstructor(c),
		tableColumns: []unit.MM{21, 7, 83, 7, 7, 10, 15, 15, 20, 13, 13, 20, 20, 8, 10, 22},
	}, nil
}

type updTemplate struct {
	constructor  *pdfego.Constructor
	dto          DTO
	tableColumns []unit.MM
}

func (tmpl *updTemplate) fill() ([]byte, error) {
	tmpl.constructor.Watermark(pdfego.WatermarkOptions{Align: "RB"}).
		Apply(tmpl.watermark)

	tmpl.constructor.Paginate(0)

	tmpl.constructor.
		Build(tmpl.header).
		Build(tmpl.tableHeader)

	tmpl.constructor.Header().
		Apply(tmpl.tableNumberHeader)

	tmpl.constructor.Repeater(tmpl.dto).
		Repeat(tmpl.tableDetails)

	tmpl.constructor.ReleaseHeader()

	tmpl.constructor.
		Build(tmpl.tableFooter).
		Build(tmpl.signatories).
		Build(tmpl.shippingHeaders).
		Build(tmpl.shippingBody)

	return tmpl.constructor.Bytes()
}

func (tmpl *updTemplate) header(block *pdfego.Block) {
	table := block.Slot().Table(3, pdfego.Columns(15, 5))

	table.Row().
		Cell("Универсальный передаточный документ", pdfego.CellOptions{Height: 10, Wrap: true, Colspan: 2, Align: "LT"})
	table.Row().
		Cell("Статус", pdfego.CellOptions{Height: 5, Align: "LM"}).
		Cell("1", pdfego.CellOptions{Height: 5, Align: "CM", Border: "O"})
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
		Blank(tmpl.dto.SfNum, "").
		Cell("от", pdfego.CellOptions{Align: "CB"}).
		Blank(tmpl.dto.SfDate, "").
		Paragraph("(1)")

	table.Row().
		Label("Исправление").
		Cell("№", pdfego.CellOptions{Align: "CB"}).
		BlankEmpty().
		Cell("от", pdfego.CellOptions{Align: "CB"}).
		BlankEmpty().
		Paragraph("(1а)")

	table = numbersDates.Slot().
		Table(1, []unit.MM{190})

	table.Row().
		Cell("Приложение № 1 к постановлению Правительства Российской Федерации от 26 декабря 2011 г. № 1137\n(в редакции постановления Правительства Российской Федерации от 16 августа 2024 г. № 1096)",
			pdfego.CellOptions{Align: "RT", FontSize: 5, Wrap: true})

	requisites := numberDatesRequisitesSlot.Block(
		pdfego.NodeOptions{
			IndentH: 1,
			IndentV: 5,
		},
	)

	requisitesColumns := []unit.MM{45, 80, 10}

	table = requisites.Slot().
		Table(10, requisitesColumns)

	table.Row().
		LabelHead("Продавец:").
		Form(tmpl.dto.OrgPrintName, "LB").
		Paragraph("(2)")
	table.Row().
		Label("Адрес:").
		Form(tmpl.dto.OrgPrintAddress, "LB").
		Paragraph("(2а)")
	table.Row().
		Label("ИНН/КПП продавца:").
		Blank(tmpl.dto.OrgInnKpp, "LB").
		Paragraph("(2б)")
	table.Row().
		Label("Грузоотправитель и его адрес:").
		Form(tmpl.dto.ShipperPrintNameAddress, "LB").
		Paragraph("(3)")
	table.Row().
		Label("Грузополучатель и его адрес:").
		Form(tmpl.dto.ConsigneePrintNameAddress, "LB").
		Paragraph("(4)")
	table.Row().
		Label("К платежно-расчетному документу №:").
		Blank(tmpl.dto.PaymentAndSettlementDocument, "LB").
		Paragraph("(5)")
	table.Row().
		Label("Документ об отгрузке:").
		Form(tmpl.dto.ShippingDocuments, "LB").
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
		Form(tmpl.dto.SupplierPrintName, "LB").
		Paragraph("(6)")
	table.Row().
		Label("Адрес:").
		Form(tmpl.dto.SupplierPrintAddress, "LB").
		Paragraph("(6а)")
	table.Row().
		Label("ИНН/КПП покупателя:").
		Blank(tmpl.dto.SupplierInnKpp, "LB").
		Paragraph("(6б)")
	table.Row().
		Label("Валюта: наименование, код:").
		Blank(tmpl.dto.CurrencyNameCode, "LB").
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

func (tmpl *updTemplate) tableDetails(block *pdfego.Block, section []string) {
	table := block.Slot().
		Table(1, tmpl.tableColumns)

	table.Row().
		Cell(section[0], pdfego.CellOptions{ID: 0, Align: "CB", Border: "tblR"}).
		Outlined(section[1], pdfego.CellOptions{ID: 1, Align: "CB"}).
		Outlined(section[2], pdfego.CellOptions{ID: 2, Align: "LB", Wrap: true}).
		Outlined(section[3], pdfego.CellOptions{ID: 3, Align: "CB"}).
		Outlined(section[4], pdfego.CellOptions{ID: 4, Align: "RB"}).
		Outlined(section[5], pdfego.CellOptions{ID: 5, Align: "LB"}).
		Outlined(section[6], pdfego.CellOptions{ID: 6, Align: "RB"}).
		Outlined(section[7], pdfego.CellOptions{ID: 7, Align: "RB"}).
		Outlined(section[8], pdfego.CellOptions{ID: 8, Align: "RB"}).
		Outlined(section[9], pdfego.CellOptions{ID: 9, Align: "RB"}).
		Outlined(section[10], pdfego.CellOptions{ID: 10, Align: "RB"}).
		Outlined(section[11], pdfego.CellOptions{ID: 11, Align: "RB"}).
		Outlined(section[12], pdfego.CellOptions{ID: 12, Align: "RB"}).
		Outlined(section[13], pdfego.CellOptions{ID: 13, Align: "RB"}).
		Outlined(section[14], pdfego.CellOptions{ID: 14, Align: "LB"}).
		Outlined(section[15], pdfego.CellOptions{ID: 15, Align: "LB"})
}

func (tmpl *updTemplate) tableFooter(block *pdfego.Block) {
	table := block.Slot().
		Table(1, tmpl.tableColumns)

	table.Row().
		Cell("", pdfego.CellOptions{Border: "tblR", Height: 3}).
		Outlined("Всего к оплате:", pdfego.CellOptions{Align: "RB", Colspan: 7, Font: pdfego.FontBold}).
		Outlined(tmpl.dto.AmountWithoutVatTotal, pdfego.CellOptions{Align: "RB"}).
		Outlined("X", pdfego.CellOptions{Align: "CB", Colspan: 2, Font: pdfego.FontBold}).
		Outlined(tmpl.dto.AmountVatTotal, pdfego.CellOptions{Align: "RB"}).
		Outlined(tmpl.dto.AmountWithVatTotal, pdfego.CellOptions{Align: "RB"}).
		Outlined("", pdfego.CellOptions{Colspan: 3})
}

func (tmpl *updTemplate) signatories(block *pdfego.Block) {
	table := block.Slot(pdfego.NodeOptions{
		IndentH: 21,
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
		Blank(tmpl.dto.OrgChiefName, "CB").
		Cell("Главный бухгалтер\nили иное уполномоченное лицо", pdfego.CellOptions{Height: 10, Align: "LB", Wrap: true}).
		Blank("[электронная подпись]", "CB").
		Blank(tmpl.dto.OrgAccountantName, "CB")
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
		Blank(tmpl.dto.Contract, "LB").
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
		Blank(tmpl.dto.StoreKeeperPosition, "").
		Blank("[электронная подпись]", "").
		Blank(tmpl.dto.StoreKeeperName, "").
		Paragraph("[12]")
	table.Row().
		Underscore("(должность)").
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)")
	table.Row(rowOptions).
		Label("Дата отгрузки, передачи (сдачи)").
		BlankSpan(tmpl.dto.DocSendDate, "", 2).
		Paragraph("[13]")
	table.Row(rowOptions).
		LabelSpan("Иные сведения об отгрузке, передаче", 2)
	table.Row(rowOptions).
		BlankSpan(tmpl.dto.LkID, "LB", 3).
		Paragraph("[14]")
	table.Row().
		UnderscoreSpan("(ссылки на неотъемлемые приложения, сопутствующие документы, иные документы и т.п.)", 3)
	table.Row(rowOptions).
		LabelSpan("Ответственный за правильность оформления факта хозяйственной жизни", 2)
	table.Row(rowOptions).
		Blank(tmpl.dto.SenderChiefPosition, "").
		Blank("[электронная подпись]", "").
		Blank(tmpl.dto.SenderChiefName, "").
		Paragraph("[15]")
	table.Row().
		Underscore("(должность)").
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)")
	table.Row(rowOptions).
		LabelSpan("Наименование экономического субъекта – составителя документа (в т.ч. комиссионера / агента)", 2)
	table.Row(rowOptions).
		BlankSpan(tmpl.dto.OrgPrintName, "LB", 3).
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
		Blank(tmpl.dto.RecipientPosition, "").
		Image("fisher", pdfego.CellOptions{Border: "b", Align: "BC", Scale: 2, OffsetV: 2, PlaceHolder: "Jeliy Fisher"}).
		Blank(tmpl.dto.RecipientName, "").
		Paragraph("[17]")
	table.Row().
		Underscore("(должность)").
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)")
	table.Row(rowOptions).
		Label("Дата получения (приемки)").
		BlankSpan(tmpl.dto.DocReceiveDate, "", 2).
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
		Blank(tmpl.dto.RecipientChiefPosition, "").
		Image("gaben", pdfego.CellOptions{Border: "b", Align: "BC", Scale: 2, OffsetV: 2, PlaceHolder: "Gabe Newell"}).
		Blank(tmpl.dto.RecipientChiefName, "").
		Paragraph("[20]")
	table.Row().
		Underscore("(должность)").
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)")
	table.Row(rowOptions).
		LabelSpan("Наименование экономического субъекта – составителя документа (в т.ч. комиссионера / агента)", 2)
	table.Row(rowOptions).
		BlankSpan(tmpl.dto.SupplierPrintName, "LB", 3).
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
		Label("номер КЭП"+tmpl.dto.Certificate.KEP, opts).
		Label(tmpl.dto.Certificate.SignDate, opts)
	table.Row().
		Skip().
		Label(tmpl.dto.Certificate.SignerName, opts).
		Label(fmt.Sprintf("период действия с %s\nпо %s", tmpl.dto.Certificate.ValidityPeriodFrom, tmpl.dto.Certificate.ValidityPeriodTo), pdfego.CellOptions{FontSize: 5, Wrap: true})
}
