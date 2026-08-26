package image

import (
	"github.com/eugene-static/pdfego/internal/components/object"
	"github.com/eugene-static/pdfego/internal/components/parameter"
)

type objects struct {
	xImage *object.Object
	xAlpha *object.Object
}

type xImage struct {
	Type             parameter.Name       `pdf:"Type"`
	Subtype          parameter.Name       `pdf:"Subtype"`
	Width            parameter.Integer    `pdf:"Width"`
	Height           parameter.Integer    `pdf:"Height"`
	ColorSpace       parameter.Name       `pdf:"ColorSpace"`
	BitsPerComponent parameter.Integer    `pdf:"BitsPerComponent"`
	Filter           parameter.Name       `pdf:"Filter"`
	Length           parameter.Integer    `pdf:"Length"`
	SMask            *parameter.Reference `pdf:"SMask,omitempty"`
}

func (img *Image) XImage() *object.Object {
	return img.objects.xImage
}

func (img *Image) initXImage(rgb []byte) (err error) {
	bytes, err := img.compressor.ForceCompress(rgb)
	if err != nil {
		return err
	}

	var xAlphaObjectNumber *parameter.Reference

	if img.objects.xAlpha != nil {
		xAlphaObjectNumber = img.objects.xAlpha.Reference()
	}

	dict := xImage{
		Type:             "XObject",
		Subtype:          "Image",
		Width:            parameter.Integer(img.width),
		Height:           parameter.Integer(img.height),
		ColorSpace:       "DeviceRGB",
		BitsPerComponent: bitsPerComponent,
		Filter:           parameter.FlateDecode,
		Length:           parameter.Integer(len(bytes)),
		SMask:            xAlphaObjectNumber,
	}

	img.objects.xImage = object.Read(dict, bytes)

	return nil
}

func (img *Image) initXAlpha(alpha []byte) (err error) {
	bytes, err := img.compressor.ForceCompress(alpha)
	if err != nil {
		return err
	}

	dict := xImage{
		Type:             "XObject",
		Subtype:          "Image",
		Width:            parameter.Integer(img.width),
		Height:           parameter.Integer(img.height),
		ColorSpace:       "DeviceGray",
		BitsPerComponent: bitsPerComponent,
		Filter:           parameter.FlateDecode,
		Length:           parameter.Integer(len(bytes)),
	}

	img.objects.xAlpha = object.Read(dict, bytes)

	return nil
}

func (img *Image) XAlpha() *object.Object {
	return img.objects.xAlpha
}
