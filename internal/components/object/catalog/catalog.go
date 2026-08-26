package catalog

import (
	"github.com/eugene-static/pdfego/internal/components/object"
	"github.com/eugene-static/pdfego/internal/components/object/pages"
	"github.com/eugene-static/pdfego/internal/components/parameter"
)

type Catalog struct {
	Type  parameter.Name      `pdf:"Type"`
	Pages parameter.Reference `pdf:"Pages"`
}

func New() *Catalog {
	return &Catalog{
		Type:  "Catalog",
		Pages: pages.ObjectNumber,
	}
}

func (cat *Catalog) Catalog() *object.Object {
	obj := object.NewSimple()

	obj.ReadDictionary(cat)

	return obj
}
