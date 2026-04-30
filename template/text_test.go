package template

import (
	"os"
	"testing"

	"github.com/golang/freetype/truetype"
	font_face "golang.org/x/image/font"
)

func Test_SplitText(t *testing.T) {
	fontBytes, err := os.ReadFile("../fonts/NotoSerifSC-Regular.ttf")
	if err != nil {
		t.Fatal(err)
	}

	fontParsed, err := truetype.Parse(fontBytes)
	if err != nil {
		t.Fatal(err)
	}

	fontFace := truetype.NewFace(fontParsed, &truetype.Options{
		Size:    6,
		DPI:     72,
		Hinting: font_face.HintingFull,
	})

	c := cell{
		width: 20,
		text:  "Универсальный\nпередаточный документ",
	}

	lines := c.splitText(fontFace)

	for _, line := range lines {
		t.Log(line)
	}

	wantLen := 3
	gotLen := len(lines)

	if gotLen != wantLen {
		t.Errorf("len(lines) = %d, want %d", gotLen, wantLen)
	}
}
