package main

import (
	"github.com/eugene-static/pdf-craft/core"
)

func (upd UPD) FillTemplate() ([]byte, error) {
	tmpl := core.New(core.Landscape)

	return output.Bytes(), nil
}
