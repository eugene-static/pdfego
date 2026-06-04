package upd

import (
	"github.com/eugene-static/pdf-craft"
)

func prepareTemplate() (*pdf_craft.Core, error) {
	c := pdf_craft.NewCore(pdf_craft.Landscape)
	c.SetMargin(3)
	c.SetDefaultFontSize(6)
	c.SetDefaultBorderSize(0.3)
	//c.Compress()

	err := c.SetFontRegular("../fonts/LiberationSans-Regular.ttf")
	if err != nil {
		return nil, err
	}

	err = c.SetFontBold("../fonts/LiberationSans-Bold.ttf")
	if err != nil {
		return nil, err
	}

	return c, nil
}
