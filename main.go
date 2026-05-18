package main

import (
	"log"
	"os"
	"time"
)

func main() {
	t := time.Now()

	upd := NewUPD(1)

	bytes, err := upd.FillTemplate()
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
