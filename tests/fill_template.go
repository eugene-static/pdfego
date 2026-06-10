package tests

import (
	"github.com/eugene-static/pdf-craft"
	"github.com/eugene-static/pdf-craft/pkg/unit"
)

func prepareTemplate() (*pdf_craft.Core, error) {
	c := pdf_craft.NewCore(pdf_craft.Landscape)
	c.SetMargins(3, 3, 3, 8)
	c.SetDefaultFontSize(6)
	c.SetDefaultBorderSize(0.3)
	c.WithCompression()

	err := c.SetFontRegular("../fonts/LiberationSans-Regular.ttf")
	if err != nil {
		return nil, err
	}

	err = c.SetFontBold("../fonts/LiberationSans-Bold.ttf")
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (dto DTO) fillTemplate(c *pdf_craft.Core) ([]byte, error) {
	err := c.ReadImage("1691194.png", "1691194")
	if err != nil {
		return nil, err
	}

	constructor := pdf_craft.New(c)

	tableColumns := []unit.MM{21, 7, 83, 7, 7, 10, 15, 15, 20, 13, 13, 20, 20, 8, 10, 22}

	constructor.Watermark(pdf_craft.WatermarkOptions{Align: "RB"}).
		Apply(dto.watermark)

	constructor.Paginate(0)

	constructor.
		Build(dto.header).
		Build(tableHeader(tableColumns))

	constructor.Header().
		Apply(tableNumberHeader(tableColumns))

	constructor.Repeater(dto).
		Repeat(tableDetails(tableColumns))

	constructor.ReleaseHeader()

	constructor.
		Build(dto.tableFooter(tableColumns)).
		Build(dto.signatories).
		Build(dto.shippingHeaders).
		Build(dto.shippingBody)

	return constructor.Bytes()
}

func (dto DTO) header(block *pdf_craft.Block) {
	table := block.Slot().Table(3, pdf_craft.Columns(15, 5))

	table.Row().
		Cell("Универсальный передаточный документ", pdf_craft.CellOptions{Height: 10, Wrap: true, Colspan: 2, Align: "LT"})
	table.Row().
		Cell("Статус", pdf_craft.CellOptions{Height: 5, Align: "LM"}).
		Cell("1", pdf_craft.CellOptions{Height: 5, Align: "CM", Border: "O"})
	table.Row().
		Cell("1 - счет-фактура и\nпередаточный\nдокумент (акт)\n2 - передаточный\nдокумент (акт)", pdf_craft.CellOptions{Height: 15, Align: "LB", Colspan: 2, FontSize: 5, Wrap: true})

	numberDatesRequisitesSlot := block.Slot(
		pdf_craft.NodeOptions{
			IndentH: 1,
			Ledge:   5,
			Border:  "L",
		},
	)

	numbersDates := numberDatesRequisitesSlot.Block(
		pdf_craft.NodeOptions{
			IndentH: 1,
		},
	)

	table = numbersDates.Slot().
		Table(2, []unit.MM{15, 5, 20, 5, 20, 10})

	table.Row().
		Label("Счет-фактура").
		Cell("№", pdf_craft.CellOptions{Align: "CB"}).
		Blank(dto.SfNum, "").
		Cell("от", pdf_craft.CellOptions{Align: "CB"}).
		Blank(dto.SfDate, "").
		Paragraph("(1)")

	table.Row().
		Label("Исправление").
		Cell("№", pdf_craft.CellOptions{Align: "CB"}).
		BlankEmpty().
		Cell("от", pdf_craft.CellOptions{Align: "CB"}).
		BlankEmpty().
		Paragraph("(1а)")

	table = numbersDates.Slot().
		Table(1, []unit.MM{190})

	table.Row().
		Cell("Приложение № 1 к постановлению Правительства Российской Федерации от 26 декабря 2011 г. № 1137\n(в редакции постановления Правительства Российской Федерации от 16 августа 2024 г. № 1096)",
			pdf_craft.CellOptions{Align: "RT", FontSize: 5, Wrap: true})

	requisites := numberDatesRequisitesSlot.Block(
		pdf_craft.NodeOptions{
			IndentH: 1,
			IndentV: 5,
		},
	)

	requisitesColumns := []unit.MM{45, 80, 10}

	table = requisites.Slot().
		Table(10, requisitesColumns)

	table.Row().
		LabelHead("Продавец:").
		Form(dto.OrgPrintName, "LB").
		Paragraph("(2)")
	table.Row().
		Label("Адрес:").
		Form(dto.OrgPrintAddress, "LB").
		Paragraph("(2а)")
	table.Row().
		Label("ИНН/КПП продавца:").
		Blank(dto.OrgInnKpp, "LB").
		Paragraph("(2б)")
	table.Row().
		Label("Грузоотправитель и его адрес:").
		Form(dto.ShipperPrintNameAddress, "LB").
		Paragraph("(3)")
	table.Row().
		Label("Грузополучатель и его адрес:").
		Form(dto.ConsigneePrintNameAddress, "LB").
		Paragraph("(4)")
	table.Row().
		Label("К платежно-расчетному документу №:").
		Blank(dto.PaymentAndSettlementDocument, "LB").
		Paragraph("(5)")
	table.Row().
		Label("Документ об отгрузке:").
		Form(dto.ShippingDocuments, "LB").
		Paragraph("(5а)")
	table.Row().
		LabelSpan(
			"К счету-фактуре (счетам-фактурам), выставленному (выставленным)\n"+
				"при получении оплаты, частичной оплаты или иных платежей в счет\n"+
				"предстоящих поставок товаров (выполнения работ, оказания услуг),",
			2,
			pdf_craft.CellOptions{Wrap: true},
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
		Form(dto.SupplierPrintName, "LB").
		Paragraph("(6)")
	table.Row().
		Label("Адрес:").
		Form(dto.SupplierPrintAddress, "LB").
		Paragraph("(6а)")
	table.Row().
		Label("ИНН/КПП покупателя:").
		Blank(dto.SupplierInnKpp, "LB").
		Paragraph("(6б)")
	table.Row().
		Label("Валюта: наименование, код:").
		Blank(dto.CurrencyNameCode, "LB").
		Paragraph("(7)")
	table.Row().
		Label("Идентификатор государственного контракта,\nдоговора (соглашения) (при наличии):", pdf_craft.CellOptions{Wrap: true}).
		BlankEmpty().
		Paragraph("(8)")
}

func tableHeader(columns []unit.MM) func(block *pdf_craft.Block) {
	return func(block *pdf_craft.Block) {
		table := block.Slot().Table(2, columns)

		optsBounded := pdf_craft.CellOptions{
			Height: 15,
			Border: "o",
			Wrap:   true,
		}

		optsBoundedRS2 := pdf_craft.CellOptions{
			Rowspan: 2,
			Border:  "o",
			Wrap:    true,
		}

		optsBoundedCS2 := pdf_craft.CellOptions{
			Colspan: 2,
			Border:  "o",
			Wrap:    true,
		}

		table.Row().
			Cell("Код\nтовара/работ, услуг", pdf_craft.CellOptions{Height: 10, Border: "tblR", Wrap: true, Rowspan: 2}).
			Cell("№\nп/п", optsBoundedRS2).
			Cell("Наименование товара\n(описание выполненных работ, оказанных услуг),\nимущественного права", optsBoundedRS2).
			Cell("Код\nвида\nтовара", optsBoundedRS2).
			Cell("Единица\nизмерения", optsBoundedCS2).
			Cell("Количество\n(объем)", optsBoundedRS2).
			Cell("Цена\n(тариф) за\nединицу\nизмерения", optsBoundedRS2).
			Cell("Стоимость\nтоваров (работ,\nуслуг),\nимущественных\nправ без налога -\nвсего", optsBoundedRS2).
			Cell("В том числе\nсумма\nакциза", optsBoundedRS2).
			Cell("Налоговая\nставка", optsBoundedRS2).
			Cell("Сумма налога,\nпредъявляемая\nпокупателю", optsBoundedRS2).
			Cell("Стоимость\nтоваров (работ,\nуслуг),\nимущественных\nправ с налогом -\nвсего", optsBoundedRS2).
			Cell("Страна\nпроисхождения\nтовара", optsBoundedCS2).
			Cell("Регистрационный\nномер декларации\nна товары или\nрегистрационный\nномер партии\nтовара,\nподлежащего\nпрослеживаемости", optsBoundedRS2)

		table.Row().
			Cell("код", optsBounded).
			Cell("условное\nобозна-\nчение\n(нацио-\nнальное)", optsBounded).
			Cell("цифро-\nвой\nкод", optsBounded).
			Cell("краткое\nнаимено-\nвание", optsBounded)
	}
}

func tableNumberHeader(columns []unit.MM) func(*pdf_craft.Block) {
	return func(block *pdf_craft.Block) {
		table := block.Slot().
			Table(1, columns)

		table.Row().
			Cell("А", pdf_craft.CellOptions{Align: "CM", Height: 3, Border: "tblR"}).
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
}

func tableDetails(columns []unit.MM) pdf_craft.RepeaterApplier {
	return func(block *pdf_craft.Block, section []string) {
		table := block.Slot().
			Table(1, columns)

		table.Row().
			Cell(section[0], pdf_craft.CellOptions{ID: 0, Align: "CB", Border: "tblR"}).
			Outlined(section[1], pdf_craft.CellOptions{ID: 1, Align: "CB"}).
			Outlined(section[2], pdf_craft.CellOptions{ID: 2, Align: "LB", Wrap: true}).
			Outlined(section[3], pdf_craft.CellOptions{ID: 3, Align: "CB"}).
			Outlined(section[4], pdf_craft.CellOptions{ID: 4, Align: "RB"}).
			Outlined(section[5], pdf_craft.CellOptions{ID: 5, Align: "LB"}).
			Outlined(section[6], pdf_craft.CellOptions{ID: 6, Align: "RB"}).
			Outlined(section[7], pdf_craft.CellOptions{ID: 7, Align: "RB"}).
			Outlined(section[8], pdf_craft.CellOptions{ID: 8, Align: "RB"}).
			Outlined(section[9], pdf_craft.CellOptions{ID: 9, Align: "RB"}).
			Outlined(section[10], pdf_craft.CellOptions{ID: 10, Align: "RB"}).
			Outlined(section[11], pdf_craft.CellOptions{ID: 11, Align: "RB"}).
			Outlined(section[12], pdf_craft.CellOptions{ID: 12, Align: "RB"}).
			Outlined(section[13], pdf_craft.CellOptions{ID: 13, Align: "RB"}).
			Outlined(section[14], pdf_craft.CellOptions{ID: 14, Align: "LB"}).
			Outlined(section[15], pdf_craft.CellOptions{ID: 15, Align: "LB"})
	}
}

func (dto DTO) tableFooter(columns []unit.MM) pdf_craft.BlockApplier {
	return func(block *pdf_craft.Block) {
		table := block.Slot().
			Table(1, columns)

		table.Row().
			Cell("", pdf_craft.CellOptions{Border: "tblR", Height: 3}).
			Outlined("Всего к оплате:", pdf_craft.CellOptions{Align: "RB", Colspan: 7, Font: pdf_craft.FontBold}).
			Outlined(dto.AmountWithoutVatTotal, pdf_craft.CellOptions{Align: "RB"}).
			Outlined("X", pdf_craft.CellOptions{Align: "CB", Colspan: 2, Font: pdf_craft.FontBold}).
			Outlined(dto.AmountVatTotal, pdf_craft.CellOptions{Align: "RB"}).
			Outlined(dto.AmountWithVatTotal, pdf_craft.CellOptions{Align: "RB"}).
			Outlined("", pdf_craft.CellOptions{Colspan: 3})
	}
}

func (dto DTO) signatories(block *pdf_craft.Block) {
	table := block.Slot(pdf_craft.NodeOptions{
		IndentH: 21,
		Border:  "LB",
		Ledge:   3,
	}).
		Table(
			4,
			[]unit.MM{38, 47, 47, 38, 47, 47},
			pdf_craft.TableOptions{SpacingH: 1, IndentH: 1})

	table.Row().
		Cell("Руководитель организации\nили иное уполномоченное лицо", pdf_craft.CellOptions{Height: 10, Align: "LB", Wrap: true}).
		Image("1691194", pdf_craft.CellOptions{Align: "BC", Border: "b", PlaceHolder: "[электронная подпись]", OffsetV: 3}).
		Blank(dto.OrgChiefName, "CB").
		Cell("Главный бухгалтер\nили иное уполномоченное лицо", pdf_craft.CellOptions{Height: 10, Align: "LB", Wrap: true}).
		Blank("[электронная подпись]", "CB").
		Blank(dto.OrgAccountantName, "CB")
	table.Row().
		Skip().
		Underscore("подпись").
		Underscore("(Ф.И.О.)").
		Skip().
		Underscore("подпись").
		Underscore("(Ф.И.О.)")
	table.Row().
		Cell("Индивидуальный предприниматель\nили иное уполномоченное лицо", pdf_craft.CellOptions{Height: 10, Align: "LB", Wrap: true}).
		BlankEmpty().
		BlankEmpty().
		BlankEmpty(pdf_craft.CellOptions{Colspan: 3})
	table.Row().
		Skip().
		Underscore("подпись").
		Underscore("(Ф.И.О.)").
		UnderscoreSpan("(основной государственный регистрационный номер индивидуального предпринимателя и дата присвоения такого номера)", 3)
}

func (dto DTO) shippingHeaders(block *pdf_craft.Block) {
	table := block.Slot().
		Table(4, []unit.MM{60, 224, 10},
			pdf_craft.TableOptions{
				IndentV: 5,
			},
		)

	table.Row().
		Label("Основание передачи (сдачи) / получения (приемки)").
		Blank(dto.Contract, "LB").
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

func (dto DTO) shippingBody(block *pdf_craft.Block) {
	columns := []unit.MM{44, 44, 44, 10}
	rowOptions := pdf_craft.RowOptions{MinHeight: 4}

	table := block.Slot(
		pdf_craft.NodeOptions{
			IndentV: 3,
		},
	).
		Table(14, columns,
			pdf_craft.TableOptions{
				SpacingH: 1,
			},
		)

	table.Row().
		LabelSpan("Товар (груз) передал / услуги, результаты работ, права сдал", 2)
	table.Row(rowOptions).
		Blank(dto.SenderChiefPosition, "").
		BlankEmpty().
		Blank(dto.SenderChiefName, "").
		Paragraph("[12]")
	table.Row().
		Underscore("(должность)").
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)")
	table.Row(rowOptions).
		Label("Дата отгрузки, передачи (сдачи)").
		BlankSpan(dto.DocSendDate, "", 2).
		Paragraph("[13]")
	table.Row(rowOptions).
		LabelSpan("Иные сведения об отгрузке, передаче", 2)
	table.Row(rowOptions).
		BlankSpan(dto.LkID, "LB", 3).
		Paragraph("[14]")
	table.Row().
		UnderscoreSpan("(ссылки на неотъемлемые приложения, сопутствующие документы, иные документы и т.п.)", 3)
	table.Row(rowOptions).
		LabelSpan("Ответственный за правильность оформления факта хозяйственной жизни", 2)
	table.Row(rowOptions).
		Blank(dto.SenderChiefPosition, "").
		BlankEmpty().
		Blank(dto.SenderChiefName, "").
		Paragraph("[15]")
	table.Row().
		Underscore("(должность)").
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)")
	table.Row(rowOptions).
		LabelSpan("Наименование экономического субъекта – составителя документа (в т.ч. комиссионера / агента)", 2)
	table.Row(rowOptions).
		BlankSpan(dto.OrgPrintName, "LB", 3).
		Paragraph("[16]")
	table.Row().
		UnderscoreSpan("(может не заполняться при проставлении печати в М.П., может быть указан ИНН / КПП)", 3)
	table.Row().
		Cell("М.П.")

	table = block.Slot(
		pdf_craft.NodeOptions{
			IndentV: 3,
			IndentH: 1,
			Border:  "L",
		},
	).
		Table(14, columns,
			pdf_craft.TableOptions{
				SpacingH: 1,
				IndentH:  3,
			},
		)

	table.Row().
		LabelSpan("Товар (груз) получил / услуги, результаты работ, права принял", 2)
	table.Row(rowOptions).
		Blank(dto.RecipientPosition, "").
		BlankEmpty().
		Blank(dto.RecipientName, "").
		Paragraph("[17]")
	table.Row().
		Underscore("(должность)").
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)")
	table.Row(rowOptions).
		Label("Дата получения (приемки)").
		BlankSpan(dto.DocReceiveDate, "", 2).
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
		Blank(dto.RecipientChiefPosition, "").
		BlankEmpty().
		Blank(dto.RecipientChiefName, "").
		Paragraph("[20]")
	table.Row().
		Underscore("(должность)").
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)")
	table.Row(rowOptions).
		LabelSpan("Наименование экономического субъекта – составителя документа (в т.ч. комиссионера / агента)", 2)
	table.Row(rowOptions).
		BlankSpan(dto.SupplierPrintName, "LB", 3).
		Paragraph("[21]")
	table.Row().
		UnderscoreSpan("(может не заполняться при проставлении печати в М.П., может быть указан ИНН / КПП)", 3)
	table.Row().
		Cell("М.П.")
}

func (dto DTO) watermark(block *pdf_craft.Block) {
	table := block.Slot().
		Table(3, []unit.MM{40, 40, 40, 40},
			pdf_craft.TableOptions{
				SpacingH:    1,
				SpacingV:    1,
				Border:      "O",
				TextColor:   pdf_craft.ColorDarkBlue,
				BorderColor: pdf_craft.ColorDarkBlue,
			})

	opts := pdf_craft.CellOptions{
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
		Label("номер КЭП 90379e6a254d4df79c93", opts).
		Label("01.06.2026б 05:45", opts)
	table.Row().
		Skip().
		Label("Сарыкова Наталья Викторовна", opts).
		Label("период действия с 19.08.2025 09:16\nпо 19.08.2026 09:26", pdf_craft.CellOptions{FontSize: 5, Wrap: true})
}
