package image

import (
	"bytes"
	"errors"
	"image"
	"image/png"
)

type Image struct {
	alias     string
	rgbData   []byte
	alphaData []byte
	width     int
	height    int
	rgbComp   bool
	alphaComp bool
	hasAlpha  bool
}

func New(alias string, data []byte) (*Image, error) {
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

	return &Image{
		alias:     alias,
		rgbData:   rgb,
		alphaData: alpha,
		width:     width,
		height:    height,
		hasAlpha:  hasAlpha,
	}, nil
}

func (img *Image) Alias() string {
	return img.alias
}

func (img *Image) RGB() ([]byte, bool) {
	return img.rgbData, img.rgbComp
}

func (img *Image) Alpha() ([]byte, bool) {
	return img.alphaData, img.alphaComp
}

func (img *Image) Width() int {
	return img.width
}

func (img *Image) Height() int {
	return img.height
}

func (img *Image) SaveCompressedRGB(data []byte) {
	img.rgbData = data

	img.rgbComp = true
}

func (img *Image) SaveCompressedAlpha(data []byte) {
	img.alphaData = data

	img.alphaComp = true
}
