package template

import (
	"testing"

	"codeberg.org/go-pdf/fpdf"
)

func TestTemplate(t *testing.T) {
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
		//core: core,
	}

	_ = tmpl
}

func (t *Template) TitleHeader(field *Field) {
	table := field.Table(15, 5)

	r1 := table.Row()
	r1.Cell("Универсальный\nпередаточный\nдокумент", CellOpts{Height: 20, Wrap: true, Colspan: 2, Align: "LT"})
	r2 := table.Row()
	r2.Cell("Статус", CellOpts{Height: 4, Align: "LC"})
	r2.Cell("1", CellOpts{Height: 5, Align: "LC"})
	r3 := table.Row()
	r3.Cell("1 - счет-фактура и\nпередаточный\nдокумент (акт)\n2 - передаточный\nдокумент (акт)", CellOpts{Align: "LB", Colspan: 2, FontSize: 5, Wrap: true})
}
