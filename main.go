package main

import (
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/eugene-static/pdf-craft/core"
	"github.com/eugene-static/pdf-craft/upd"
)

func main() {
	t := time.Now()

	c := core.New(core.Landscape)
	c.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	c.SetMargin(3)
	c.SetDefaultFontSize(6)
	c.SetBorders(0.2, 1)

	err := c.SetFontRegular("./fonts/LiberationSans-Regular.ttf")
	if err != nil {
		log.Fatal(err)
	}

	err = c.SetFontBold("./fonts/LiberationSans-Bold.ttf")
	if err != nil {
		log.Fatal(err)
	}

	upd := upd.NewUPD(10000)

	bytes, err := upd.FillTemplate(c)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("duration: %v\n", time.Since(t))

	output, err := os.Create("output.pdf")
	if err != nil {
		log.Fatal(err)
	}

	defer output.Close()

	_, err = output.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}
}
