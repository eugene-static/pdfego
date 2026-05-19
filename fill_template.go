package main

import (
	"log/slog"
	"os"

	"github.com/eugene-static/pdf-craft/core"
	tmpl "github.com/eugene-static/pdf-craft/template"
)

func (upd UPD) FillTemplate() ([]byte, error) {
	c := core.New(core.Landscape)
	c.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	c.SetMargin(3)
	c.SetDefaultFontSize(6)
	c.SetBorders(0.2, 0.8)

	err := c.SetFontRegular("./fonts/LiberationSans-Regular.ttf")
	if err != nil {
		return nil, err
	}

	err = c.SetFontBold("./fonts/LiberationSans-Bold.ttf")
	if err != nil {
		return nil, err
	}

	t := tmpl.New(c)

	headBlock := t.Block(tmpl.Options{Spacing: 3})

	headBlock.
		Slot().
		Add(upd.titleHeader)

	headBlock.
		Slot().
		Add(upd.numberHeader).
		Add(upd.requisites)

	t.
		Block(tmpl.Options{Indent: 5}).
		Slot().
		Add(tableHeader)

	t.Render()

	return t.Bytes(), nil
}

func (upd UPD) titleHeader(slot *tmpl.Slot) {
	table := slot.Table(15, 5)

	r1 := table.Row()
	r1.Cell("Универсальный передаточный документ", tmpl.CellOpts{Height: 10, Wrap: true, Colspan: 2, Align: "LT"})
	r2 := table.Row()
	r2.Cell("Статус", tmpl.CellOpts{Height: 4, Align: "LM"})
	r2.Cell("1", tmpl.CellOpts{Height: 5, Align: "CM", Border: "O"})
	r3 := table.Row()
	r3.Cell("1 - счет-фактура и\nпередаточный\nдокумент (акт)\n2 - передаточный\nдокумент (акт)", tmpl.CellOpts{Height: 10, Align: "LB", Colspan: 2, FontSize: 5, Wrap: true})
}

func (upd UPD) numberHeader(slot *tmpl.Slot) {
	numbersDates := slot.Block()
	numbers := numbersDates.Slot()

	table := numbers.Table(15, 5, 20, 5, 20, 10)

	r1 := table.Row()
	r1.Label("Счет-фактура")
	r1.Cell("№", tmpl.CellOpts{Align: "CB"})
	r1.FormC(upd.SfNum, false)
	r1.Cell("от", tmpl.CellOpts{Align: "CB"})
	r1.FormC(upd.SfDate, false)
	r1.Paragraph("(1)")

	r2 := table.Row()
	r2.Label("Исправление")
	r2.Cell("№", tmpl.CellOpts{Align: "CB"})
	r2.FormC("", false)
	r2.Cell("от", tmpl.CellOpts{Align: "CB"})
	r2.FormC("", false)
	r2.Paragraph("(1а)")

	edition := numbersDates.Slot()

	table = edition.Table(190)
	r1 = table.Row()
	r1.Cell("Приложение № 1 к постановлению Правительства Российской Федерации от 26 декабря 2011 г. № 1137\n(в редакции постановления Правительства Российской Федерации от 16 августа 2024 г. № 1096)",
		tmpl.CellOpts{Align: "RT", FontSize: 5, Wrap: true})
}

func (upd UPD) requisites(slot *tmpl.Slot) {
	requisites := slot.Block(tmpl.Options{Indent: 5})
	company := requisites.Slot()

	table := company.Table(45, 80, 10)

	table.Row().LabelHead("Продавец:").FormL(upd.OrgPrintName, true).Paragraph("(2)")
	table.Row().Label("Адрес:").FormL(upd.OrgPrintAddress, true).Paragraph("(2а)")
	table.Row().Label("ИНН/КПП продавца:").FormL(upd.OrgInnKpp, false).Paragraph("(2б)")
	table.Row().Label("Грузоотправитель и его адрес:").FormL(upd.ShipperPrintNameAddress, false).Paragraph("(3)")
	table.Row().Label("Грузополучатель и его адрес:").FormL(upd.ConsigneePrintNameAddress, false).Paragraph("(4)")
	table.Row().Label("К платежно-расчетному документу №:").FormL(upd.PaymentAndSettlementDocument, false).Paragraph("(5)")
	table.Row().Label("Документ об отгрузке:").FormL(upd.ShippingDocuments, true).Paragraph("(5а)")
	table.Row().Cell(
		"К счету-фактуре (счетам-фактурам), выставленному (выставленным)\nпри получении оплаты, частичной оплаты или иных платежей в счет\nпредстоящих поставок товаров (выполнения работ, оказания услуг),",
		tmpl.CellOpts{Align: "LB", Wrap: true, Colspan: 2})
	table.Row().Label("передачи имущественных прав №:").FormL("", false)
	table.Row().Label("исправление №:").FormL("", false).Paragraph("(5б)")

	supplier := requisites.Slot()
	table = supplier.Table(45, 80, 10)

	table.Row().LabelHead("Покупатель:").FormL(upd.SupplierPrintName, true).Paragraph("(6)")
	table.Row().Label("Адрес:").FormL(upd.SupplierPrintAddress, true).Paragraph("(6а)")
	table.Row().Label("ИНН/КПП покупателя:").FormL(upd.SupplierInnKpp, false).Paragraph("(6б)")
	table.Row().Label("Валюта: наименование, код:").FormL(upd.CurrencyNameCode, false).Paragraph("(7)")
	table.Row().
		Cell("Идентификатор государственного контракта,\nдоговора (соглашения) (при наличии):", tmpl.CellOpts{Align: "LB", Wrap: true}).
		FormL("", false).
		Paragraph("(8)")
}

func tableHeader(slot *tmpl.Slot) {
	table := slot.Table(21, 5, 40, 9, 10, 10, 20, 20, 25, 20, 20, 25, 25, 10, 10, 25)

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
		Cell("Код\nтовара/работ, услуг", optsBoundedRS2).
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
		Cell("цифровой\nкод", optsBounded).
		Cell("краткое\nнаимено-\nвание", optsBounded)
}
