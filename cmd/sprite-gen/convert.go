package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"

	"github.com/GyeongHoKim/pokemon-cli/internal/sprite/spritedata"
)

// maxDim caps the longer side of a converted sprite in pixels (never
// upscaled). Tuned by eyeballing `cmd/sprite-gen -preview` output.
const maxDim = 40

// alphaTrimThreshold is the alpha value below which a pixel is considered
// empty for the purposes of trimming the sprite's transparent border.
const alphaTrimThreshold = 8

// targetFrameCount is how many frames each species keeps after subsampling
// its native animation (typically ~40-50 native frames — full embedding
// would bloat the binary and outpace any sane terminal redraw rate).
const targetFrameCount = 8

// decodeAnimatedSprite decodes an animated GIF into a subsampled sequence of
// PixelGrids. GIF frames are disposal-aware deltas against a persistent
// canvas, not standalone images — see compositeGIFFrames.
func decodeAnimatedSprite(data []byte) ([]spritedata.PixelGrid, error) {
	g, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode gif: %w", err)
	}

	composited := compositeGIFFrames(g)

	grids := make([]spritedata.PixelGrid, 0, len(composited))
	for _, frame := range composited {
		trimmed := trimTransparentBorder(frame, alphaTrimThreshold)
		grid, err := toPixelGrid(trimmed)
		if err != nil {
			return nil, fmt.Errorf("frame %d: %w", len(grids), err)
		}
		grids = append(grids, grid)
	}

	return subsampleFrames(grids, targetFrameCount), nil
}

// compositeGIFFrames replays a decoded GIF's disposal-method deltas onto a
// persistent canvas and returns one full-canvas snapshot per native frame,
// in order. GIF frames are typically small sub-rectangles that must be
// drawn onto (and, depending on disposal, cleared from) an accumulating
// canvas — rendering g.Image[i] in isolation produces cropped/wrong output.
func compositeGIFFrames(g *gif.GIF) []*image.RGBA {
	bounds := image.Rect(0, 0, g.Config.Width, g.Config.Height)
	canvas := image.NewRGBA(bounds)
	// These sprites rely on GIF frame transparency, not a background color
	// index, so the canvas starts (and, on disposal 2, resets to) fully
	// transparent rather than any GIF-declared background color.

	frames := make([]*image.RGBA, len(g.Image))
	var preDrawSnapshot *image.RGBA

	for i, srcFrame := range g.Image {
		if g.Disposal[i] == gif.DisposalPrevious {
			preDrawSnapshot = image.NewRGBA(bounds)
			draw.Draw(preDrawSnapshot, bounds, canvas, bounds.Min, draw.Src)
		}

		draw.Draw(canvas, srcFrame.Bounds(), srcFrame, srcFrame.Bounds().Min, draw.Over)

		snapshot := image.NewRGBA(bounds)
		draw.Draw(snapshot, bounds, canvas, bounds.Min, draw.Src)
		frames[i] = snapshot

		switch g.Disposal[i] {
		case gif.DisposalBackground:
			draw.Draw(canvas, srcFrame.Bounds(), image.Transparent, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			draw.Draw(canvas, bounds, preDrawSnapshot, bounds.Min, draw.Src)
		default: // DisposalNone / unspecified: leave canvas as-is, next frame draws on top
		}
	}
	return frames
}

// subsampleFrames picks up to n evenly-spaced frames from frames, preserving
// order. If frames has n or fewer elements, it's returned unchanged.
func subsampleFrames(frames []spritedata.PixelGrid, n int) []spritedata.PixelGrid {
	if len(frames) <= n {
		return frames
	}
	step := len(frames) / n
	out := make([]spritedata.PixelGrid, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, frames[i*step])
	}
	return out
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
