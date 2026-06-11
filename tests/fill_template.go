package tests

import (
	"fmt"

	"github.com/eugene-static/pdf-craft"
	"github.com/eugene-static/pdf-craft/pkg/unit"
)

func prepareTemplate() (*updTemplate, error) {
	c := pdf_craft.NewCore(pdf_craft.Landscape)
	c.SetMargins(3, 3, 3, 8)
	c.SetDefaultFontSize(6)
	c.SetDefaultBorderSize(0.3)
	c.EnableCompression()
	c.IgnoreImageNotFound()

	err := c.SetFontRegular("../fonts/LiberationSans-Regular.ttf")
	if err != nil {
		return nil, err
	}

	err = c.SetFontBold("../fonts/LiberationSans-Bold.ttf")
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
		constructor:  pdf_craft.New(c),
		tableColumns: []unit.MM{21, 7, 83, 7, 7, 10, 15, 15, 20, 13, 13, 20, 20, 8, 10, 22},
	}, nil
}

type updTemplate struct {
	constructor  *pdf_craft.Constructor
	dto          DTO
	tableColumns []unit.MM
}

func (tmpl *updTemplate) fill() ([]byte, error) {
	tmpl.constructor.Watermark(pdf_craft.WatermarkOptions{Align: "RB"}).
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

func (tmpl *updTemplate) header(block *pdf_craft.Block) {
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
		Blank(tmpl.dto.SfNum, "").
		Cell("от", pdf_craft.CellOptions{Align: "CB"}).
		Blank(tmpl.dto.SfDate, "").
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
		Label("Идентификатор государственного контракта,\nдоговора (соглашения) (при наличии):", pdf_craft.CellOptions{Wrap: true}).
		BlankEmpty().
		Paragraph("(8)")
}

func (tmpl *updTemplate) tableHeader(block *pdf_craft.Block) {
	table := block.Slot().Table(2, tmpl.tableColumns)

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

func (tmpl *updTemplate) tableNumberHeader(block *pdf_craft.Block) {
	table := block.Slot().
		Table(1, tmpl.tableColumns)

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

func (tmpl *updTemplate) tableDetails(block *pdf_craft.Block, section []string) {
	table := block.Slot().
		Table(1, tmpl.tableColumns)

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

func (tmpl *updTemplate) tableFooter(block *pdf_craft.Block) {
	table := block.Slot().
		Table(1, tmpl.tableColumns)

	table.Row().
		Cell("", pdf_craft.CellOptions{Border: "tblR", Height: 3}).
		Outlined("Всего к оплате:", pdf_craft.CellOptions{Align: "RB", Colspan: 7, Font: pdf_craft.FontBold}).
		Outlined(tmpl.dto.AmountWithoutVatTotal, pdf_craft.CellOptions{Align: "RB"}).
		Outlined("X", pdf_craft.CellOptions{Align: "CB", Colspan: 2, Font: pdf_craft.FontBold}).
		Outlined(tmpl.dto.AmountVatTotal, pdf_craft.CellOptions{Align: "RB"}).
		Outlined(tmpl.dto.AmountWithVatTotal, pdf_craft.CellOptions{Align: "RB"}).
		Outlined("", pdf_craft.CellOptions{Colspan: 3})
}

func (tmpl *updTemplate) signatories(block *pdf_craft.Block) {
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
		Blank("[электронная подпись]", "CB").
		Blank(tmpl.dto.OrgChiefName, "CB").
		Cell("Главный бухгалтер\nили иное уполномоченное лицо", pdf_craft.CellOptions{Height: 10, Align: "LB", Wrap: true}).
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
		Cell("Индивидуальный предприниматель\nили иное уполномоченное лицо", pdf_craft.CellOptions{Height: 10, Align: "LB", Wrap: true}).
		BlankEmpty().
		BlankEmpty().
		BlankEmpty(pdf_craft.CellOptions{Colspan: 3})
	table.Row().
		Skip().
		Underscore("(подпись)").
		Underscore("(Ф.И.О.)").
		UnderscoreSpan("(основной государственный регистрационный номер индивидуального предпринимателя и дата присвоения такого номера)", 3)
}

func (tmpl *updTemplate) shippingHeaders(block *pdf_craft.Block) {
	table := block.Slot().
		Table(4, []unit.MM{60, 224, 10},
			pdf_craft.TableOptions{
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

func (tmpl *updTemplate) shippingBody(block *pdf_craft.Block) {
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
		Blank(tmpl.dto.RecipientPosition, "").
		Image("fisher", pdf_craft.CellOptions{Border: "b", Align: "BC", Scale: 2, OffsetV: 2, PlaceHolder: "Jeliy Fisher"}).
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
		Image("gaben", pdf_craft.CellOptions{Border: "b", Align: "BC", Scale: 2, OffsetV: 2, PlaceHolder: "Gabe Newell"}).
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

func (tmpl *updTemplate) watermark(block *pdf_craft.Block) {
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
		Label("номер КЭП"+tmpl.dto.Certificate.KEP, opts).
		Label(tmpl.dto.Certificate.SignDate, opts)
	table.Row().
		Skip().
		Label(tmpl.dto.Certificate.SignerName, opts).
		Label(fmt.Sprintf("период действия с %s\nпо %s", tmpl.dto.Certificate.ValidityPeriodFrom, tmpl.dto.Certificate.ValidityPeriodTo), pdf_craft.CellOptions{FontSize: 5, Wrap: true})
}
