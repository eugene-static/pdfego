package main

import (
	"fmt"
	"os"
	"time"
)

func main() {

	t := time.Now()

	printForm, err := upd.FillTemplate()
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Printf("duration: %s", time.Since(t).String())

	output, err := os.Create("output.pdf")
	if err != nil {
		fmt.Println(err)

		return
	}

	defer output.Close()

	_, err = output.Write(printForm)
	if err != nil {
		fmt.Println(err)

		return
	}
}
