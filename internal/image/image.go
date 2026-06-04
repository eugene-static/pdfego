package image

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

type Image struct {
	alias           string
	rgbData         []byte
	rgbCompressed   []byte
	alphaData       []byte
	alphaCompressed []byte
	width           int
	height          int
	colored         bool
	hasAlpha        bool
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

	rgb := make([]byte, pixCount*3)
	var alpha []byte
	if hasAlpha {
		alpha = make([]byte, pixCount)
	}

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
		colored:   true,
		hasAlpha:  hasAlpha,
	}, nil
}

func (img *Image) Alias() string {
	return img.alias
}

func (img *Image) RGB() []byte {
	return img.rgbData
}

func (img *Image) RGBCompressed() ([]byte, bool) {
	return img.rgbCompressed, img.rgbCompressed != nil
}

func (img *Image) Alpha() []byte {
	return img.alphaData
}

func (img *Image) AlphaCompressed() ([]byte, bool) {
	return img.alphaCompressed, img.alphaCompressed != nil
}

func (img *Image) Width() int {
	return img.width
}

func (img *Image) Height() int {
	return img.height
}

func (img *Image) ColorSpace() string {
	if img.colored {
		return "RGB"
	}

	return "Gray"
}

func (img *Image) SaveCompressedRGB(data []byte) {
	img.rgbCompressed = data
}

func (img *Image) SaveCompressedAlpha(data []byte) {
	img.alphaCompressed = data
}

// Удаляет альфа-канал из RGBA (просто отбрасывает байты A).
func convertRGBAToRGB(rgba []byte) []byte {
	rgb := make([]byte, len(rgba)/4*3)
	for i, j := 0, 0; i < len(rgba); i, j = i+4, j+3 {
		copy(rgb[j:j+3], rgba[i:i+3])
	}
	return rgb
}

func convertToRGB(bounds image.Rectangle, colFunc func(x, y int) color.Color) []byte {
	rgba := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, colFunc(x, y))
		}
	}

	return convertRGBAToRGB(rgba.Pix)
}
