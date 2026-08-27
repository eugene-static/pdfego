package dictionary

import "github.com/eugene-static/pdfego/internal/components/parameter"

type Content struct {
	Filter parameter.Name    `pdf:"Filter,omitempty"`
	Length parameter.Integer `pdf:"Length"`
}
