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
		core: core,
	}

	_ = tmpl
}
