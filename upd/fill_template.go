package upd

import (
	"github.com/eugene-static/pdf-craft"
	"github.com/eugene-static/pdf-craft/pkg/unit"
)

func (upd UPD) fillTemplate(c *pdf_craft.Core) ([]byte, error) {
	template := pdf_craft.New(c)

	headBlock := template.Block()

	headBlock.Slot().
		Add(upd.titleHeader)

	headBlock.Slot(
		pdf_craft.NodeOptions{
			IndentLeft: 1,
			Ledge:      5,
			Border:     "L",
		},
	).
		Add(upd.numberHeader).
		Add(upd.requisites)

	tableColumns := []unit.MM{21, 7, 83, 7, 7, 10, 15, 15, 20, 13, 13, 20, 20, 8, 10, 22}

	template.Block().
		Add(tableHeader(tableColumns))

	template.Header().
		Add(tableNumberHeader(tableColumns))

	template.Repeater(upd).
		AddRepeater(tableDetails(tableColumns))

	template.Render()

	return template.Bytes()
}

func (upd UPD) titleHeader(slot *pdf_craft.Slot) {
	table := slot.Table(3, []unit.MM{15, 5})

	table.Row().
		Cell("Универсальный передаточный документ", pdf_craft.CellOptions{Height: 10, Wrap: true, Colspan: 2, Align: "LT"})
	table.Row().
		Cell("Статус", pdf_craft.CellOptions{Height: 5, Align: "LM"}).
		Cell("1", pdf_craft.CellOptions{Height: 5, Align: "CM", Border: "O"})
	table.Row().
		Cell("1 - счет-фактура и\nпередаточный\nдокумент (акт)\n2 - передаточный\nдокумент (акт)", pdf_craft.CellOptions{Height: 15, Align: "LB", Colspan: 2, FontSize: 5, Wrap: true})
}

func (upd UPD) numberHeader(slot *pdf_craft.Slot) {
	numbersDates := slot.Block(
		pdf_craft.NodeOptions{
			IndentLeft: 1,
		},
	)

	table := numbersDates.Slot().
		Table(2, []unit.MM{15, 5, 20, 5, 20, 10})

	table.Row().
		Label("Счет-фактура").
		Cell("№", pdf_craft.CellOptions{Align: "CB"}).
		FormC(upd.SfNum, false).
		Cell("от", pdf_craft.CellOptions{Align: "CB"}).
		FormC(upd.SfDate, false).
		Paragraph("(1)")

	table.Row().
		Label("Исправление").
		Cell("№", pdf_craft.CellOptions{Align: "CB"}).
		FormC("", false).
		Cell("от", pdf_craft.CellOptions{Align: "CB"}).
		FormC("", false).
		Paragraph("(1а)")

	table = numbersDates.Slot().
		Table(1, []unit.MM{190})

	table.Row().
		Cell("Приложение № 1 к постановлению Правительства Российской Федерации от 26 декабря 2011 г. № 1137\n(в редакции постановления Правительства Российской Федерации от 16 августа 2024 г. № 1096)",
			pdf_craft.CellOptions{Align: "RT", FontSize: 5, Wrap: true})
}

func (upd UPD) requisites(slot *pdf_craft.Slot) {
	requisites := slot.Block(
		pdf_craft.NodeOptions{
			IndentLeft: 1,
			IndentTop:  5,
		},
	)

	table := requisites.Slot().
		Table(10, []unit.MM{45, 80, 10})

	table.Row().
		LabelHead("Продавец:").
		FormL(upd.OrgPrintName, true).
		Paragraph("(2)")
	table.Row().
		Label("Адрес:").
		FormL(upd.OrgPrintAddress, true).
		Paragraph("(2а)")
	table.Row().
		Label("ИНН/КПП продавца:").
		FormL(upd.OrgInnKpp, false).
		Paragraph("(2б)")
	table.Row().
		Label("Грузоотправитель и его адрес:").
		FormL(upd.ShipperPrintNameAddress, false).
		Paragraph("(3)")
	table.Row().
		Label("Грузополучатель и его адрес:").
		FormL(upd.ConsigneePrintNameAddress, false).
		Paragraph("(4)")
	table.Row().
		Label("К платежно-расчетному документу №:").
		FormL(upd.PaymentAndSettlementDocument, false).
		Paragraph("(5)")
	table.Row().
		Label("Документ об отгрузке:").
		FormL(upd.ShippingDocuments, true).
		Paragraph("(5а)")
	table.Row().
		Cell(
			"К счету-фактуре (счетам-фактурам), выставленному (выставленным)\nпри получении оплаты, частичной оплаты или иных платежей в счет\nпредстоящих поставок товаров (выполнения работ, оказания услуг),",
			pdf_craft.CellOptions{
				Align:   "LB",
				Wrap:    true,
				Colspan: 2,
			},
		)
	table.Row().
		Label("передачи имущественных прав №:").
		FormL("", false)
	table.Row().
		Label("исправление №:").
		FormL("", false).
		Paragraph("(5б)")

	table = requisites.Slot().
		Table(5, []unit.MM{45, 80, 10})

	table.Row().
		LabelHead("Покупатель:").
		FormL(upd.SupplierPrintName, true).
		Paragraph("(6)")
	table.Row().
		Label("Адрес:").
		FormL(upd.SupplierPrintAddress, true).
		Paragraph("(6а)")
	table.Row().
		Label("ИНН/КПП покупателя:").
		FormL(upd.SupplierInnKpp, false).
		Paragraph("(6б)")
	table.Row().
		Label("Валюта: наименование, код:").
		FormL(upd.CurrencyNameCode, false).
		Paragraph("(7)")
	table.Row().
		Cell("Идентификатор государственного контракта,\nдоговора (соглашения) (при наличии):", pdf_craft.CellOptions{Align: "LB", Wrap: true}).
		FormL("", false).
		Paragraph("(8)")
}

func tableHeader(columns []unit.MM) func(block *pdf_craft.Block) {
	return func(block *pdf_craft.Block) {
		table := block.Slot().Table(2, columns)

		optsBounded := pdf_craft.CellOptions{
			Height: 15,
			Align:  "CM",
			Border: "o",
			Wrap:   true,
		}

		optsBoundedRS2 := pdf_craft.CellOptions{
			Border:  "o",
			Wrap:    true,
			Rowspan: 2,
		}

		optsBoundedCS2 := pdf_craft.CellOptions{
			Align:   "CM",
			Border:  "o",
			Wrap:    true,
			Colspan: 2,
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
			Bounded("1").
			Bounded("1а").
			Bounded("1б").
			Bounded("2").
			Bounded("2а").
			Bounded("3").
			Bounded("4").
			Bounded("5").
			Bounded("6").
			Bounded("7").
			Bounded("8").
			Bounded("9").
			Bounded("10").
			Bounded("10а").
			Bounded("11")
	}
}

func tableDetails(columns []unit.MM) pdf_craft.RepeaterFiller {
	return func(block *pdf_craft.Block, ordered []string) {
		table := block.Slot().
			Table(1, columns)

		table.Row().
			Cell(ordered[0], pdf_craft.CellOptions{ID: 0, Align: "CB", Border: "tblR"}).
			Cell(ordered[1], pdf_craft.CellOptions{ID: 1, Align: "CB", Border: "o"}).
			Cell(ordered[2], pdf_craft.CellOptions{ID: 2, Align: "LB", Border: "o", Wrap: true}).
			Cell(ordered[3], pdf_craft.CellOptions{ID: 3, Align: "CB", Border: "o"}).
			Cell(ordered[4], pdf_craft.CellOptions{ID: 4, Align: "RB", Border: "o"}).
			Cell(ordered[5], pdf_craft.CellOptions{ID: 5, Align: "LB", Border: "o"}).
			Cell(ordered[6], pdf_craft.CellOptions{ID: 6, Align: "RB", Border: "o"}).
			Cell(ordered[7], pdf_craft.CellOptions{ID: 7, Align: "RB", Border: "o"}).
			Cell(ordered[8], pdf_craft.CellOptions{ID: 8, Align: "RB", Border: "o"}).
			Cell(ordered[9], pdf_craft.CellOptions{ID: 9, Align: "RB", Border: "o"}).
			Cell(ordered[10], pdf_craft.CellOptions{ID: 10, Align: "RB", Border: "o"}).
			Cell(ordered[11], pdf_craft.CellOptions{ID: 11, Align: "RB", Border: "o"}).
			Cell(ordered[12], pdf_craft.CellOptions{ID: 12, Align: "RB", Border: "o"}).
			Cell(ordered[13], pdf_craft.CellOptions{ID: 13, Align: "RB", Border: "o"}).
			Cell(ordered[14], pdf_craft.CellOptions{ID: 14, Align: "LB", Border: "o"}).
			Cell(ordered[15], pdf_craft.CellOptions{ID: 15, Align: "LB", Border: "o"})
	}
}
