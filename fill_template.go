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

	//err = core.SetFontBold("../fonts/LiberationSans-Bold.ttf")
	//if err != nil {
	//	t.Error(err)
	//}

	t := tmpl.New(c)

	headBlock := t.Block()
	headBlock.Slot().Add(upd.titleHeader)
	headBlock.Slot().Add(upd.numberHeader)

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

func (upd UPD) numberHeader(field *tmpl.Slot) {
	numbersDates := field.Block()
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
	r1.Cell("Приложение № 1 к постановлению\nПравительства Российской Федерации\nот 26 декабря 2011 г. № 1137\n(в редакции постановления\nПравительства Российской Федерации\nот 16 августа 2024 г. № 1096)",
		tmpl.CellOpts{Align: "RT", FontSize: 5, Wrap: true})
}
