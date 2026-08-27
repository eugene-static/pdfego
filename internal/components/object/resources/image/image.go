package image

import (
	"bytes"
	"errors"
	"image"
	"image/png"

	"github.com/eugene-static/pdfego/internal/components/parameter"
	"github.com/eugene-static/pdfego/internal/components/stream"
	"github.com/eugene-static/pdfego/internal/components/stream/primitives"
	"github.com/eugene-static/pdfego/internal/compressor"
)

const (
	bitsPerComponent = 8
)

type Image struct {
	alias      primitives.Alias
	compressor *compressor.Compressor
	objects    objects
	width      int
	height     int
}

func New(alias string, data []byte, comp *compressor.Compressor) (*Image, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	pixCount := width * height

	nrgba := image.NewRGBA(bounds)

	switch src := img.(type) {
	case *image.NRGBA:
		copy(nrgba.Pix, src.Pix)
	case *image.RGBA:
		for i := 0; i < len(src.Pix); i += 4 {
			r, g, b, a := src.Pix[i], src.Pix[i+1], src.Pix[i+2], src.Pix[i+3]
			if a != 0 && a != 255 {
				r = uint8(uint32(r) * 255 / uint32(a))
				g = uint8(uint32(g) * 255 / uint32(a))
				b = uint8(uint32(b) * 255 / uint32(a))
			}

			nrgba.Pix[i] = r
			nrgba.Pix[i+1] = g
			nrgba.Pix[i+2] = b
			nrgba.Pix[i+3] = a
		}
	default:
		return nil, errors.New("неподдерживаемый тип изображения")
	}

	hasAlpha := false

	for i := 3; i < len(nrgba.Pix); i += 4 {
		if nrgba.Pix[i] != 255 {
			hasAlpha = true

			break
		}
	}

	var alpha []byte

	if hasAlpha {
		alpha = make([]byte, pixCount)
	}

	rgb := make([]byte, pixCount*3)

	for i := range pixCount {
		off := i * 4

		copy(rgb[i*3:], nrgba.Pix[off:off+3])

		if hasAlpha {
			alpha[i] = nrgba.Pix[off+3]
		}
	}

	_image := &Image{
		alias:      primitives.Alias(alias),
		width:      width,
		height:     height,
		compressor: comp,
	}

	if hasAlpha {
		err = _image.initXAlpha(alpha)
		if err != nil {
			return nil, err
		}
	}

	err = _image.initXImage(rgb)
	if err != nil {
		return nil, err
	}

	return _image, nil
}

func (img *Image) Alias() parameter.Name {
	return parameter.Name(img.alias)
}

func (img *Image) Width() int {
	return img.width
}

func (img *Image) Height() int {
	return img.height
}

// q $W 0 0 $H $X $Y cm /$ImageAlias Do Q
func (img *Image) WriteToStream(dst *stream.Stream, matrix primitives.Matrix) {
	primitives.NewXObject(img.alias, matrix).
		WriteToStream(dst)
}
