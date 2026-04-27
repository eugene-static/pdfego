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
	core.SetLineWidth(0.1)
	core.AddPage()

	tmpl := Template{
		core: core,
	}

	tmpl.Frames(
		NewFrame().Fields(
			NewField().Drawers(
				NewTable(14, 6).Rows(
					NewRow().Cells(
						Text("Универсальный\nпередаточный\nдокумент", CellOpts{Height: 15, Align: "LT", Colspan: 2, Wrap: true}),
					),
					NewRow().Cells(
						Text("Статус", CellOpts{Height: 8, Align: "LM"}),
						Cell("1", CellOpts{Height: 4, Align: "MC", Border: "1+"}),
					),
					NewRow().Cells(
						Text("1 - счет-фактура и\nпередаточный\nдокумент (акт)\n2 - передаточный\nдокумент (акт)", CellOpts{Height: 20, Align: "LB", Colspan: 2, FontSize: 5, Wrap: true}),
					),
				),
			),
			NewField().Drawers(
				NewFrame().Fields(
					NewField().Drawers(
						NewTable(25, 30, 5, 30, 10).Rows(
							NewRow().Cells(
								Text("Счет-фактура №", CellOpts{Height: 3, Align: "LB"}),
								Form("", CellOpts{Align: "BM"}),
								Text("от", CellOpts{Align: "BM"}),
								Form("", CellOpts{Align: "BM"}),
								Text("(1)", CellOpts{Align: "BM"}),
							),
							NewRow().Cells(
								Text("Исправление №", CellOpts{Height: 3, Align: "LB"}),
								Form("", CellOpts{Align: "BM"}),
								Text("от", CellOpts{Align: "BM"}),
								Form("", CellOpts{Align: "BM"}),
								Text("(1а)", CellOpts{Align: "BM"}),
							),
						),
					),
					NewField().Drawers(
						NewTable(0).Rows(
							NewRow().Cells(
								Cell(
									"Приложение № 1 к постановлению Правительства Российской федерации от 26 декабря 2011 г. № 1137\n(в редакции постановления Правительства Российской Федерации от 16 августа Исправление № от (1а) 2024 г. № 1096)",
									CellOpts{Height: 5, Align: "RT", FontSize: 5, Wrap: true},
								),
							),
						),
					),
				),
				//NewFrame(),
			),
		),
	)

	tmpl.Draw()

	err := tmpl.core.OutputFileAndClose("output.pdf")
	if err != nil {
		t.Fatal(err)
	}
}
