package template

import (
	"log/slog"
	"os"
	"testing"

	core "github.com/eugene-static/pdf-craft/core"
)

func TestTemplate(t *testing.T) {
	core := core.New(core.Landscape)
	core.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	core.SetMargin(3)
	core.SetDefaultFontSize(6)
	core.SetBorders(0.2, 1)

	err := core.SetFontRegular("../fonts/LiberationSans-Regular.ttf")
	if err != nil {
		t.Error(err)
	}

	//err = core.SetFontBold("../fonts/LiberationSans-Bold.ttf")
	//if err != nil {
	//	t.Error(err)
	//}

	tmpl := Template{
		core: core,
	}

	headBlock := tmpl.Block()
	headBlock.Slot().Add(titleHeader)
	headBlock.Slot().Add(numberHeader)

	tmpl.Render()

	output, err := os.Create("output.pdf")
	if err != nil {
		t.Error(err)
	}

	defer output.Close()

	_, err = output.Write(tmpl.Bytes())
	if err != nil {
		t.Error(err)
	}
}

func titleHeader(slot *Slot) {
	table := slot.Table(15, 5)

	r1 := table.Row()
	r1.Cell("Универсальный передаточный документ", CellOpts{Height: 15, Wrap: true, Colspan: 2, Align: "LT"})
	r2 := table.Row()
	r2.Cell("Статус", CellOpts{Height: 4, Align: "LM"})
	r2.Cell("1", CellOpts{Height: 5, Align: "CM", Border: "O"})
	r3 := table.Row()
	r3.Cell("1 - счет-фактура и\nпередаточный\nдокумент (акт)\n2 - передаточный\nдокумент (акт)", CellOpts{Align: "LB", Colspan: 2, FontSize: 5, Wrap: true})
}

func numberHeader(field *Slot) {
	numbersDates := field.Block()
	numbers := numbersDates.Slot()

	table := numbers.Table(15, 5, 15, 5, 15, 10)

	r1 := table.Row()
	r1.Cell("Счет-фактура", CellOpts{Height: 3, Align: "LB"})
	r1.Cell("№", CellOpts{Align: "CB"})
	r1.Cell("", CellOpts{Align: "CB", Border: "b"})
	r1.Cell("от", CellOpts{Align: "CB"})
	r1.Cell("", CellOpts{Align: "CB", Border: "b"})
	r1.Cell("(1)", CellOpts{Align: "CB"})

	r2 := table.Row()
	r2.Cell("Исправление", CellOpts{Height: 3, Align: "LB"})
	r2.Cell("№", CellOpts{Align: "CB"})
	r2.Cell("", CellOpts{Align: "CB", Border: "b"})
	r2.Cell("от", CellOpts{Align: "CB"})
	r2.Cell("", CellOpts{Align: "CB", Border: "b"})
	r2.Cell("(1а)", CellOpts{Align: "CB"})
}
