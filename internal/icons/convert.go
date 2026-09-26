package icons

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/gen2brain/webp"
	"golang.org/x/image/draw"
)

const (
	iconSize    = 64
	iconQuality = 90
)

func convert(raw []byte, profile Profile) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	switch profile {
	case Icon:
		if bounds := img.Bounds(); bounds.Dx() < iconSize && bounds.Dy() < iconSize {
			return encode(img, true)
		}
		return encode(scale(img, iconSize), false)
	case Lossless:
		return encode(img, true)
	default:
		return nil, fmt.Errorf("unknown profile %q", profile)
	}
}

func scale(img image.Image, size int) image.Image {
	bounds := img.Bounds()
	if bounds.Dx() == size && bounds.Dy() == size {
		return img
	}
	scaled := image.NewNRGBA(image.Rect(0, 0, size, size))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), img, bounds, draw.Src, nil)
	return scaled
}

func encode(img image.Image, lossless bool) ([]byte, error) {
	var buf bytes.Buffer
	options := webp.Options{Quality: iconQuality, Method: 4, Lossless: lossless}
	if err := webp.Encode(&buf, img, options); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
