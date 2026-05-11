package template

import (
	"log/slog"
	"os"
	"testing"
)

func TestTemplate(t *testing.T) {
	core := New(Landscape)
	core.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	core.SetMargin(3)
	core.SetDefaultFontSize(6)
	core.SetBorders(0.2, 0.5)

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

	frame1 := tmpl.Frame()
	field1 := frame1.Field()

	tmpl.TitleHeader(field1)

	tmpl.Render()
	//
	//buf := tmpl.Buffer()
	//
	//output, err := os.Create("output.pdf")
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//defer output.Close()
	//
	//_, err = output.Write(buf.Bytes())
	//if err != nil {
	//	t.Error(err)
	//}
}

func (t *Template) TitleHeader(field *Field) {
	table := field.Table(15, 5)

	r1 := table.Row()
	r1.Cell("Универсальный передаточный документ", CellOpts{Height: 20, Wrap: true, Colspan: 2, Align: "CT", Border: "tL"})
	//r2 := table.Row()
	//r2.Cell("Статус", CellOpts{Height: 4, Align: "LC"})
	//r2.Cell("1", CellOpts{Height: 5, Align: "LC"})
	//r3 := table.Row()
	//r3.Cell("1 - счет-фактура и\nпередаточный\nдокумент (акт)\n2 - передаточный\nдокумент (акт)", CellOpts{Align: "LB", Colspan: 2, FontSize: 5, Wrap: true})
}
