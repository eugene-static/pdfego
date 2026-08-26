package manager

import (
	"github.com/eugene-static/pdfego/internal/components/object/catalog"
	"github.com/eugene-static/pdfego/internal/components/object/header"
	"github.com/eugene-static/pdfego/internal/components/object/info"
	"github.com/eugene-static/pdfego/internal/components/object/pages"
	"github.com/eugene-static/pdfego/internal/components/object/resources"
	"github.com/eugene-static/pdfego/internal/components/object/trailer"
	"github.com/eugene-static/pdfego/internal/components/object/xref"
	"github.com/eugene-static/pdfego/internal/components/stream"
	"github.com/eugene-static/pdfego/internal/components/stream/primitives"
	"github.com/eugene-static/pdfego/internal/compressor"
	"github.com/eugene-static/pdfego/unit"
)

type Manager struct {
	context    *Context
	stream     *stream.Stream
	compressor *compressor.Compressor
	header     *header.Header
	catalog    *catalog.Catalog
	resources  *resources.Resources
	pages      *pages.Pages
	info       *info.Info
	xref       *xref.XRef
	trailer    *trailer.Trailer
	error      error
}

func New(cfg *pages.Config) *Manager {
	ctx := &Context{
		hexb:  make([]primitives.HEX, 0, 1<<10),
		runeb: make([]rune, 0),
	}

	comp := compressor.New()

	return &Manager{
		context:    ctx,
		stream:     stream.New(),
		compressor: comp,
		header:     header.New(),
		catalog:    catalog.New(),
		resources:  resources.New(),
		pages:      pages.New(comp, cfg),
		info:       info.New(),
		xref:       xref.New(),
		trailer:    trailer.New(),
	}
}

func (mgr *Manager) Bytes() []byte {
	return mgr.stream.Bytes()
}

func (mgr *Manager) HeaderStream() *stream.Stream {
	return mgr.pages.HeaderStream()
}

func (mgr *Manager) WatermarkStream() *stream.Stream {
	return mgr.pages.WatermarkStream()
}

func (mgr *Manager) PageStream() *stream.Stream {
	return mgr.pages.PageStream()
}

func (mgr *Manager) Context() *Context {
	return mgr.context
}

func (mgr *Manager) Resources() *resources.Resources {
	return mgr.resources
}

func (mgr *Manager) Compressor() *compressor.Compressor {
	return mgr.compressor
}

func (mgr *Manager) Pages() *pages.Pages {
	return mgr.pages
}

func (mgr *Manager) Info() *info.Info {
	return mgr.info
}

func (mgr *Manager) NewPage() {
	mgr.pages.NewPage()
}

func (mgr *Manager) ViewBox() pages.ViewBox {
	return mgr.pages.ViewBox()
}

func (mgr *Manager) HeaderRelease() {
	mgr.pages.HeaderRelease()
}

func (mgr *Manager) EnableCompression(enable bool) {
	mgr.compressor.Enable(enable)
}

func (mgr *Manager) ResetBuffers() {
	mgr.context.hexb = mgr.context.hexb[:0]
	mgr.context.runeb = mgr.context.runeb[:0]
}

func (mgr *Manager) IsBelowBottomBorder(y unit.MM) bool {
	return mgr.pages.Config().MarginBottom+y > 0
}

type Context struct {
	hexb  []primitives.HEX
	runeb []rune
	error error
}

func (ctx *Context) HexBuffer() *[]primitives.HEX {
	return &ctx.hexb
}

func (ctx *Context) RuneBuffer() *[]rune {
	ctx.runeb = ctx.runeb[:0]
	return &ctx.runeb
}

func (ctx *Context) Error() error {
	return ctx.error
}

func (ctx *Context) SetError(err error) {
	ctx.error = err
}

func (ctx *Context) IsError() bool {
	return ctx.error != nil
}
