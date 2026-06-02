package upd

import (
	"log/slog"
	"os"

	"github.com/eugene-static/pdf-craft/internal/core"
)

func prepareTemplate() (*core.Core, error) {
	c := core.New(core.Landscape)
	c.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	c.SetMargin(3)
	c.SetDefaultFontSize(6)
	c.SetDefaultBorderSize(0.3)
	c.Compress()

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
