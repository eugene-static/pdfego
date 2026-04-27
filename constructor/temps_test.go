package constructor

import (
	"testing"

	"codeberg.org/go-pdf/fpdf"
)

func TestTmpl_Draw(t *testing.T) {
	core := fpdf.New("L", "mm", "A4", "")

	core.SetCompression(false)
	core.SetAutoPageBreak(true, 3)
	core.AddUTF8Font("NotoSerif", "", "../fonts/NotoSerifSC-Regular.ttf")
	core.AddUTF8Font("NotoSerif", "B", "../fonts/NotoSerifSC-ExtraBold.ttf")
	core.SetFont("NotoSerif", "", 6)
	core.SetMargins(3, 3, -1)
	core.AddPage()

	tmpl := Template{
		core: core,
	}

	tmpl.Grids(
		NewGrid(
			NewField(
				NewTable(
					[]float64{14, 6},
					NewRow(
						NewCell("Универсальный\nпередаточный\nдокумент", CellOpts{Height: 15, Align: "LT", Colspan: 2, Wrap: true}),
					),
					NewRow(
						NewCell("Статус", CellOpts{Height: 8, Align: "LM"}),
						NewCell("1", CellOpts{Height: 4, Align: "MC", Border: "1+"}),
					),
					NewRow(
						NewCell("1 - счет-фактура и\nпередаточный\nдокумент (акт)\n2 - передаточный\nдокумент (акт)", CellOpts{Height: 20, Align: "LB", Colspan: 2, FontSize: 5, Wrap: true}),
					),
				),
			),
			NewField(
				NewGrid(
					NewField(
						NewTable(
							[]float64{25, 30, 5, 30, 10},
							NewRow(
								NewCell("Счет-фактура №", CellOpts{Height: 3, Align: "LB"}),
								NewFormCell("", CellOpts{Align: "BM"}),
								NewCell("от", CellOpts{Align: "BM"}),
								NewFormCell("", CellOpts{Align: "BM"}),
								NewCell("(1)", CellOpts{Align: "BM"}),
							),
							NewRow(
								NewCell("Исправление №", CellOpts{Height: 3, Align: "LB"}),
								NewFormCell("", CellOpts{Align: "BM"}),
								NewCell("от", CellOpts{Align: "BM"}),
								NewFormCell("", CellOpts{Align: "BM"}),
								NewCell("(1)", CellOpts{Align: "BM"}),
							),
						),
					),
					NewField(
						NewTable(
							[]float64{0},
							NewRow(
								NewCell(
									"Приложение № 1 к постановлению Правительства Российской федерации от 26 декабря 2011 г. № 1137\n(в редакции постановления Правительства Российской Федерации от 16 августа Исправление № от (1а) 2024 г. № 1096)",
									CellOpts{Height: 5, Align: "RT", FontSize: 5, Wrap: true},
								),
							),
						),
					),
				),
			),
		),
	)

	tmpl.Draw()

	err := tmpl.core.OutputFileAndClose("output.pdf")
	if err != nil {
		t.Fatal(err)
	}
}
