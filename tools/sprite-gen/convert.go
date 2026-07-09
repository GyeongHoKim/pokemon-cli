package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/png"

	"github.com/GyeongHoKim/pokemon-cli/internal/sprite/spritedata"
)

// maxDim caps the longer side of a converted sprite in pixels (never
// upscaled). Tuned by eyeballing `just sprite-preview` output.
const maxDim = 40

// alphaTrimThreshold is the alpha value below which a pixel is considered
// empty for the purposes of trimming the sprite's transparent border.
const alphaTrimThreshold = 8

// decodeSprite decodes arbitrary PNG/GIF sprite bytes into a PixelGrid.
// Source images may be palette-indexed or full RGBA (PokeAPI/sprites mixes
// both), so this always rebuilds its own deduplicated palette rather than
// relying on the source's concrete image type.
func decodeSprite(data []byte) (spritedata.PixelGrid, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return spritedata.PixelGrid{}, fmt.Errorf("decode image: %w", err)
	}

	trimmed := trimTransparentBorder(img, alphaTrimThreshold)
	return toPixelGrid(trimmed)
}

// trimTransparentBorder returns the sub-image of img with fully-transparent
// outer rows/columns removed. If the entire image is transparent, img's
// original bounds are returned unchanged.
func trimTransparentBorder(img image.Image, threshold uint8) image.Image {
	b := img.Bounds()
	minX, minY, maxX, maxY := b.Max.X, b.Max.Y, b.Min.X, b.Min.Y
	found := false

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if uint8(a>>8) < threshold {
				continue
			}
			found = true
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	if !found {
		return img
	}

	rect := image.Rect(minX, minY, maxX+1, maxY+1)
	sub, ok := img.(interface {
		SubImage(r image.Rectangle) image.Image
	})
	if !ok {
		return img
	}
	return sub.SubImage(rect)
}

// toPixelGrid nearest-neighbor downscales img (capping the longer side at
// maxDim, never upscaling) and rebuilds it as an indexed PixelGrid with a
// freshly deduplicated palette.
func toPixelGrid(img image.Image) (spritedata.PixelGrid, error) {
	b := img.Bounds()
	origW, origH := b.Dx(), b.Dy()
	if origW == 0 || origH == 0 {
		return spritedata.PixelGrid{}, fmt.Errorf("empty image")
	}

	dstW, dstH := origW, origH
	if longer := max(origW, origH); longer > maxDim {
		scale := float64(maxDim) / float64(longer)
		dstW = max(1, int(float64(origW)*scale))
		dstH = max(1, int(float64(origH)*scale))
	}

	palette := make(map[color.RGBA]uint8)
	var paletteList []color.RGBA
	indices := make([]uint8, dstW*dstH)

	for dy := 0; dy < dstH; dy++ {
		srcY := b.Min.Y + dy*origH/dstH
		for dx := 0; dx < dstW; dx++ {
			srcX := b.Min.X + dx*origW/dstW
			r, g, bl, a := img.At(srcX, srcY).RGBA()
			c := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(bl >> 8), A: uint8(a >> 8)}
			if c.A < alphaTrimThreshold {
				c = color.RGBA{} // fully transparent, canonicalize to avoid palette bloat
			}

			idx, ok := palette[c]
			if !ok {
				if len(paletteList) >= 256 {
					return spritedata.PixelGrid{}, fmt.Errorf("palette overflow (>256 unique colors) for %dx%d sprite", dstW, dstH)
				}
				idx = uint8(len(paletteList))
				palette[c] = idx
				paletteList = append(paletteList, c)
			}
			indices[dy*dstW+dx] = idx
		}
	}

	return spritedata.PixelGrid{
		Width:   uint16(dstW),
		Height:  uint16(dstH),
		Palette: paletteList,
		Indices: indices,
	}, nil
}
