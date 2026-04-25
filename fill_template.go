package main

import (
	"bytes"
	"fmt"
)

func (upd HTML) FillTemplate() ([]byte, error) {
	tmpl := New(upd)

	output := bytes.NewBuffer(nil)
	err := tmpl.document.Output(output)
	if err != nil {
		fmt.Println(err)

		return nil, err
	}

	tmpl.document.Close()

	return output.Bytes(), nil
}
