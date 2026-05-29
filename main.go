package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/eugene-static/pdf-craft/core"
	"github.com/eugene-static/pdf-craft/upd"
)

func main() {
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

	upd := upd.NewUPD(83000)

	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	t := time.Now()

	bytes, err := upd.FillTemplate(c)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("duration: %v\n", time.Since(t))

	var memAfter runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	totalBytesAllocated := memAfter.TotalAlloc - memBefore.TotalAlloc
	totalObjectsAllocated := memAfter.Mallocs - memBefore.Mallocs

	c.Log().Debug("выделено памяти за один прогон, МБ", slog.String("total_alloc", fmt.Sprintf("%.2f", float64(totalBytesAllocated)/(1024*1024))))
	c.Log().Debug("выделено памяти в куче, МБ", slog.String("heap_in_use", fmt.Sprintf("%.2f", float64(memAfter.HeapInuse)/(1024*1024))))
	c.Log().Debug("количество аллокаций за один прогон", slog.Uint64("malloc", totalObjectsAllocated))

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
