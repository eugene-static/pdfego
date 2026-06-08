package image

import (
	"bytes"
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

	for y := range height {
		for x := range width {
			r, g, b, a := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()

			// RGBA() возвращает значения в диапазоне [0, 65535], сужаем до 8 бит
			off := nrgba.PixOffset(x, y)

			nrgba.Pix[off+0] = uint8(r >> 8)
			nrgba.Pix[off+1] = uint8(g >> 8)
			nrgba.Pix[off+2] = uint8(b >> 8)
			nrgba.Pix[off+3] = uint8(a >> 8)
		}
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
