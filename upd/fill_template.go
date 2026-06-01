package upd

import (
	"github.com/eugene-static/pdf-craft/internal/core/core"
	"github.com/eugene-static/pdf-craft/pkg/meter"
	tmpl "github.com/eugene-static/pdf-craft/template"
)

func (upd UPD) FillTemplate(c *core.Core) ([]byte, error) {
	template := tmpl.New(c)

	headBlock := template.Block(tmpl.Options{Spacing: 3})

	headBlock.
		Slot().
		Add(upd.titleHeader)

	headBlock.
		Slot().
		Add(upd.numberHeader).
		Add(upd.requisites)

	template.
		Block(
			tmpl.Options{Indent: 5},
		).
		Slot().
		Add(tableHeader)

	template.
		Header().
		Add(tableNumberHeader)

	template.
		Repeater(upd).
		Repeat(tableDetails)

	template.Render()

	return template.Bytes(), nil
}

func (upd UPD) titleHeader(slot *tmpl.Slot) {
	table := slot.Table(3, []meter.MM{15, 5})

	r1 := table.Row()
	r1.CellWithOpts("Универсальный передаточный документ", &tmpl.CellOpts{Height: 10, Wrap: true, Colspan: 2, Align: "LT"})
	r2 := table.Row()
	r2.CellWithOpts("Статус", &tmpl.CellOpts{Height: 4, Align: "LM"})
	r2.CellWithOpts("1", &tmpl.CellOpts{Height: 5, Align: "CM", Border: "O"})
	r3 := table.Row()
	r3.CellWithOpts("1 - счет-фактура и\nпередаточный\nдокумент (акт)\n2 - передаточный\nдокумент (акт)", &tmpl.CellOpts{Height: 10, Align: "LB", Colspan: 2, FontSize: 5, Wrap: true})
}

func (upd UPD) numberHeader(slot *tmpl.Slot) {
	numbersDates := slot.Block()
	numbers := numbersDates.Slot()

	table := numbers.Table(2, []meter.MM{15, 5, 20, 5, 20, 10})

	table.Row().
		Label("Счет-фактура").
		CellWithOpts("№", &tmpl.CellOpts{Align: "CB"}).
		FormC(upd.SfNum, false).
		CellWithOpts("от", &tmpl.CellOpts{Align: "CB"}).
		FormC(upd.SfDate, false).
		Paragraph("(1)")

	table.Row().
		Label("Исправление").
		CellWithOpts("№", &tmpl.CellOpts{Align: "CB"}).
		FormC("", false).
		CellWithOpts("от", &tmpl.CellOpts{Align: "CB"}).
		FormC("", false).
		Paragraph("(1а)")

	edition := numbersDates.Slot()

	table = edition.Table(1, []meter.MM{190})

	table.Row().
		CellWithOpts("Приложение № 1 к постановлению Правительства Российской Федерации от 26 декабря 2011 г. № 1137\n(в редакции постановления Правительства Российской Федерации от 16 августа 2024 г. № 1096)",
			&tmpl.CellOpts{Align: "RT", FontSize: 5, Wrap: true})
}

func (upd UPD) requisites(slot *tmpl.Slot) {
	requisites := slot.Block(
		tmpl.Options{Indent: 5},
	)
	company := requisites.Slot()

	table := company.Table(10, []meter.MM{45, 80, 10})

	table.Row().LabelHead("Продавец:").FormL(upd.OrgPrintName, true).Paragraph("(2)")
	table.Row().Label("Адрес:").FormL(upd.OrgPrintAddress, true).Paragraph("(2а)")
	table.Row().Label("ИНН/КПП продавца:").FormL(upd.OrgInnKpp, false).Paragraph("(2б)")
	table.Row().Label("Грузоотправитель и его адрес:").FormL(upd.ShipperPrintNameAddress, false).Paragraph("(3)")
	table.Row().Label("Грузополучатель и его адрес:").FormL(upd.ConsigneePrintNameAddress, false).Paragraph("(4)")
	table.Row().Label("К платежно-расчетному документу №:").FormL(upd.PaymentAndSettlementDocument, false).Paragraph("(5)")
	table.Row().Label("Документ об отгрузке:").FormL(upd.ShippingDocuments, true).Paragraph("(5а)")
	table.Row().CellWithOpts(
		"К счету-фактуре (счетам-фактурам), выставленному (выставленным)\nпри получении оплаты, частичной оплаты или иных платежей в счет\nпредстоящих поставок товаров (выполнения работ, оказания услуг),",
		&tmpl.CellOpts{Align: "LB", Wrap: true, Colspan: 2})
	table.Row().Label("передачи имущественных прав №:").FormL("", false)
	table.Row().Label("исправление №:").FormL("", false).Paragraph("(5б)")

	supplier := requisites.Slot()
	table = supplier.Table(5, []meter.MM{45, 80, 10})

	table.Row().LabelHead("Покупатель:").FormL(upd.SupplierPrintName, true).Paragraph("(6)")
	table.Row().Label("Адрес:").FormL(upd.SupplierPrintAddress, true).Paragraph("(6а)")
	table.Row().Label("ИНН/КПП покупателя:").FormL(upd.SupplierInnKpp, false).Paragraph("(6б)")
	table.Row().Label("Валюта: наименование, код:").FormL(upd.CurrencyNameCode, false).Paragraph("(7)")
	table.Row().
		CellWithOpts("Идентификатор государственного контракта,\nдоговора (соглашения) (при наличии):", &tmpl.CellOpts{Align: "LB", Wrap: true}).
		FormL("", false).
		Paragraph("(8)")
}

func tableHeader(slot *tmpl.Slot) {
	table := slot.Table(2, []meter.MM{21, 5, 83, 9, 7, 10, 15, 15, 20, 13, 13, 20, 20, 8, 10, 22})

	optsBounded := tmpl.CellOpts{
		Height: 15,
		Align:  "CM",
		Border: "o",
		Wrap:   true,
	}

	optsBoundedRS2 := tmpl.CellOpts{
		Height:  10,
		Align:   "CM",
		Border:  "o",
		Wrap:    true,
		Rowspan: 2,
	}

	optsBoundedCS2 := tmpl.CellOpts{
		Align:   "CM",
		Border:  "o",
		Wrap:    true,
		Colspan: 2,
	}

	table.Row().
		CellWithOpts("Код\nтовара/работ, услуг", &optsBoundedRS2).
		CellWithOpts("№\nп/п", &optsBoundedRS2).
		CellWithOpts("Наименование товара\n(описание выполненных работ, оказанных услуг),\nимущественного права", &optsBoundedRS2).
		CellWithOpts("Код\nвида\nтовара", &optsBoundedRS2).
		CellWithOpts("Единица\nизмерения", &optsBoundedCS2).
		CellWithOpts("Количество\n(объем)", &optsBoundedRS2).
		CellWithOpts("Цена\n(тариф) за\nединицу\nизмерения", &optsBoundedRS2).
		CellWithOpts("Стоимость\nтоваров (работ,\nуслуг),\nимущественных\nправ без налога -\nвсего", &optsBoundedRS2).
		CellWithOpts("В том числе\nсумма\nакциза", &optsBoundedRS2).
		CellWithOpts("Налоговая\nставка", &optsBoundedRS2).
		CellWithOpts("Сумма налога,\nпредъявляемая\nпокупателю", &optsBoundedRS2).
		CellWithOpts("Стоимость\nтоваров (работ,\nуслуг),\nимущественных\nправ с налогом -\nвсего", &optsBoundedRS2).
		CellWithOpts("Страна\nпроисхождения\nтовара", &optsBoundedCS2).
		CellWithOpts("Регистрационный\nномер декларации\nна товары или\nрегистрационный\nномер партии\nтовара,\nподлежащего\nпрослеживаемости", &optsBoundedRS2)

	table.Row().
		CellWithOpts("код", &optsBounded).
		CellWithOpts("условное\nобозна-\nчение\n(нацио-\nнальное)", &optsBounded).
		CellWithOpts("цифро-\nвой\nкод", &optsBounded).
		CellWithOpts("краткое\nнаимено-\nвание", &optsBounded)
}

func tableNumberHeader(block *tmpl.Block) {
	table := block.Slot().Table(1, []meter.MM{21, 5, 83, 9, 7, 10, 15, 15, 20, 13, 13, 20, 20, 8, 10, 22})

	table.Row().
		CellWithOpts("А", &tmpl.CellOpts{Align: "CM", Height: 3, Border: "tblR"}).
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

func tableDetails(block *tmpl.Block, ordered []string) {
	table := block.Slot().Table(1, []meter.MM{21, 5, 83, 9, 7, 10, 15, 15, 20, 13, 13, 20, 20, 8, 10, 22})

	table.Row().
		CellWithOpts(ordered[0], &tmpl.CellOpts{ID: 0, Align: "CB", Border: "tblR"}).
		CellWithOpts(ordered[1], &tmpl.CellOpts{ID: 1, Align: "CB", Border: "o"}).
		CellWithOpts(ordered[2], &tmpl.CellOpts{ID: 2, Align: "LB", Border: "o", Wrap: true}).
		CellWithOpts(ordered[3], &tmpl.CellOpts{ID: 3, Align: "CB", Border: "o"}).
		CellWithOpts(ordered[4], &tmpl.CellOpts{ID: 4, Align: "RB", Border: "o"}).
		CellWithOpts(ordered[5], &tmpl.CellOpts{ID: 5, Align: "LB", Border: "o"}).
		CellWithOpts(ordered[6], &tmpl.CellOpts{ID: 6, Align: "RB", Border: "o"}).
		CellWithOpts(ordered[7], &tmpl.CellOpts{ID: 7, Align: "RB", Border: "o"}).
		CellWithOpts(ordered[8], &tmpl.CellOpts{ID: 8, Align: "RB", Border: "o"}).
		CellWithOpts(ordered[9], &tmpl.CellOpts{ID: 9, Align: "RB", Border: "o"}).
		CellWithOpts(ordered[10], &tmpl.CellOpts{ID: 10, Align: "RB", Border: "o"}).
		CellWithOpts(ordered[11], &tmpl.CellOpts{ID: 11, Align: "RB", Border: "o"}).
		CellWithOpts(ordered[12], &tmpl.CellOpts{ID: 12, Align: "RB", Border: "o"}).
		CellWithOpts(ordered[13], &tmpl.CellOpts{ID: 13, Align: "RB", Border: "o"}).
		CellWithOpts(ordered[14], &tmpl.CellOpts{ID: 14, Align: "LB", Border: "o"}).
		CellWithOpts(ordered[15], &tmpl.CellOpts{ID: 15, Align: "LB", Border: "o"})
}
