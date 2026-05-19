package upd

import (
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/eugene-static/pdf-craft/core"
)

func BenchmarkUPD_FillTemplate(b *testing.B) {
	upd := NewUPD(83000)

	c := core.New(core.Landscape)
	c.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	c.SetMargin(3)
	c.SetDefaultFontSize(6)
	c.SetBorders(0.2, 1)

	err := c.SetFontRegular("../fonts/LiberationSans-Regular.ttf")
	if err != nil {
		log.Fatal(err)
	}

	err = c.SetFontBold("../fonts/LiberationSans-Bold.ttf")
	if err != nil {
		log.Fatal(err)
	}

	b.ReportAllocs()

	for range b.N {
		_, err = upd.FillTemplate(c)
		if err != nil {
			b.Fatal(err)
		}
	}
}
