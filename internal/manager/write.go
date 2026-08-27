package manager

import (
	"github.com/eugene-static/pdfego/internal/components/dictionary"
	"github.com/eugene-static/pdfego/internal/components/object"
	"github.com/eugene-static/pdfego/internal/components/object/resources"
	"github.com/eugene-static/pdfego/internal/components/object/resources/font"
	"github.com/eugene-static/pdfego/internal/components/object/resources/image"
	"github.com/eugene-static/pdfego/internal/components/parameter"
)

func (mgr *Manager) WritePage() {
	if mgr.context.IsError() {
		return
	}

	content, err := mgr.pages.PageContent()
	if err != nil {
		mgr.context.SetError(err)

		return
	}

	mgr.writeObject(content)

	page := mgr.pages.Page()

	mgr.writeObject(page)
}

func (mgr *Manager) WriteHeader() {
	if mgr.context.IsError() {
		return
	}

	header, err := mgr.pages.HeaderContent()
	if err != nil {
		mgr.context.SetError(err)

		return
	}

	mgr.writeObject(header)
}

func (mgr *Manager) WriteWatermark() {
	if mgr.context.IsError() {
		return
	}

	watermark, err := mgr.pages.WatermarkContent()
	if err != nil {
		mgr.context.SetError(err)

		return
	}

	mgr.writeObject(watermark)
}

func (mgr *Manager) StartDocument() {
	mgr.header.WriteToStream(mgr.stream)
}

func (mgr *Manager) FinishDocument(resources *resources.Resources) {
	mgr.writeResources(resources)
	mgr.writePages()
	mgr.writeInfo()
	mgr.writeCatalog()
	mgr.writeXref()
	mgr.writeTrailer()
}

func (mgr *Manager) newXrefRecord() parameter.Reference {
	xLen := mgr.stream.Len()

	return mgr.xref.NewRecord(uint(xLen))
}

func (mgr *Manager) updateXrefRecord(objNum parameter.Reference) {
	xLen := mgr.stream.Len()

	err := mgr.xref.UpdateRecord(objNum, uint(xLen))
	if err != nil {
		mgr.context.SetError(err)

		return
	}
}

func (mgr *Manager) writeResources(resources *resources.Resources) {
	if mgr.context.IsError() {
		return
	}

	fontsDictionary := dictionary.New()
	xObjectsDictionary := dictionary.New()

	resources.ForEachFont(func(alias string, _font *font.Font) {
		number := mgr.writeFont(_font)

		fontsDictionary.Set(parameter.Name(alias), number)
	})

	resources.ForEachImage(func(alias string, _image *image.Image) {
		number := mgr.writeImage(_image)

		xObjectsDictionary.Set(parameter.Name(alias), number)
	})

	mgr.resources.ForEachFont(func(alias string, _font *font.Font) {
		number := mgr.writeFont(_font)

		fontsDictionary.Set(parameter.Name(alias), number)
	})

	mgr.resources.ForEachImage(func(alias string, _image *image.Image) {
		number := mgr.writeImage(_image)

		xObjectsDictionary.Set(parameter.Name(alias), number)
	})

	mgr.resources.ForEachObject(func(alias string, _object *object.Object) {
		mgr.writeObject(_object)

		xObjectsDictionary.Set(parameter.Name(alias), _object.Number())
	})

	res := resources.Resources(fontsDictionary, xObjectsDictionary)

	mgr.updateXrefRecord(res.Number())

	res.WriteToStream(mgr.stream)
}

func (mgr *Manager) writeFont(f *font.Font) parameter.Reference {
	if mgr.context.IsError() {
		return 0
	}

	fontFile2, err := f.FontFile2()
	if err != nil {
		mgr.context.SetError(err)

		return 0
	}

	cMapB := f.CMapB()
	fontDescriptor := f.FontDescriptor()
	cidFontType2 := f.CIDFontType2()
	fontType0 := f.FontType0()

	mgr.writeObject(cMapB)
	mgr.writeObject(fontFile2)
	mgr.writeObject(fontDescriptor)
	mgr.writeObject(cidFontType2)
	mgr.writeObject(fontType0)

	return fontType0.Number()
}

func (mgr *Manager) writeImage(img *image.Image) parameter.Reference {
	if mgr.context.IsError() {
		return 0
	}

	xAlpha := img.XAlpha()
	xImage := img.XImage()

	mgr.writeObject(xAlpha)
	mgr.writeObject(xImage)

	return xImage.Number()
}

func (mgr *Manager) writePages() {
	if mgr.context.IsError() {
		return
	}

	pages := mgr.pages.Pages()

	mgr.updateXrefRecord(pages.Number())

	pages.WriteToStream(mgr.stream)
}

func (mgr *Manager) writeInfo() {
	if mgr.context.IsError() {
		return
	}

	info := mgr.info.Info()

	mgr.writeObject(info)

	mgr.trailer.SetInfo(info.Number())
}

func (mgr *Manager) writeCatalog() {
	if mgr.context.IsError() {
		return
	}

	catalog := mgr.catalog.Catalog()

	mgr.writeObject(catalog)

	mgr.trailer.SetRoot(catalog.Number())
}

func (mgr *Manager) writeXref() {
	if mgr.context.IsError() {
		return
	}

	mgr.xref.UpdateBytesOffset(uint64(mgr.stream.Len()))
	mgr.xref.WriteToStream(mgr.stream)
}

func (mgr *Manager) writeTrailer() {
	if mgr.context.IsError() {
		return
	}

	mgr.trailer.SetSize(mgr.xref.RecordsCount())

	mgr.trailer.WriteToStream(mgr.stream, mgr.xref.Index())
}

func (mgr *Manager) writeObject(obj *object.Object) {
	if mgr.context.IsError() || obj == nil {
		return
	}

	obj.SetNumber(mgr.newXrefRecord())
	obj.WriteToStream(mgr.stream)
}
