package main

import (
	"log"
	"log/slog"
	"os"
	"runtime"
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

	upd := upd.NewUPD(2)

	var m runtime.MemStats

	bytes, err := upd.FillTemplate(c)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("duration: %v\n", time.Since(t))

	runtime.ReadMemStats(&m)
	peakMegabytes := float64(m.HeapInuse) / (1024 * 1024)

	log.Printf("[INFO] Пиковое потребление памяти в куче: %.2f MB\n", peakMegabytes)

	output, err := os.Create("output.pdf")
	if err != nil {
		log.Fatal(err)
	}

	defer output.Close()

	outputText, err := os.Create("output_text.txt")
	if err != nil {
		log.Fatal(err)
	}

	defer outputText.Close()

	_, err = output.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}

	_, err = outputText.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}
}
