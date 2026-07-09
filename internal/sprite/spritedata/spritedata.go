// Package spritedata defines the wire format shared between tools/sprite-gen
// (which produces it) and internal/sprite (which embeds and decodes it).
package spritedata

import "image/color"

// PixelGrid is a row-major indexed pixel grid: Indices[y*Width+x] indexes
// into Palette. Palette entries carry their own alpha, so transparency is
// determined per-pixel via the referenced palette entry's A channel.
type PixelGrid struct {
	Width, Height uint16
	Palette       []color.RGBA
	Indices       []uint8
}

// SpeciesSprites holds the two poses rendered for a single species.
type SpeciesSprites struct {
	Icon   PixelGrid
	Battle PixelGrid
}

// Bundle is the full embedded dataset, keyed by national Pokédex id.
type Bundle map[int]SpeciesSprites
