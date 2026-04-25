package tempst

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

	tmpl := Tmpl{
		core: core,
	}

	tmpl.Grids(
		NewGrid(
			2,
			1,
		).
			Cells(
				NewGridCell().
					Tables(
						NewTable(
							vals(14, 6),
						).
							Rows(
								NewRow().Cells(
									NewTableCell(TableCellOpts{Height: 15, Align: "J", Colspan: 2}).SetText(Text{wrap: true, body: "Универсальный\nпередаточный\nдокумент"}),
								),
								NewRow().Cells(
									NewTableCell(TableCellOpts{Height: 4, Align: "LM"}).SetString("Статус"),
									NewTableCell(TableCellOpts{Height: 4, Align: "MC", Border: "1+"}).SetString("1"),
								),
								NewRow().Cells(
									NewTableCell(TableCellOpts{Height: 20, Align: "LB", Colspan: 2}).SetText(Text{wrap: true, fontSize: 5, body: "1 - счет-фактура и\nпередаточный\nдокумент (акт)\n2 - передаточный\nдокумент (акт)"}),
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
