// Package sprite resolves Pokémon species to terminal-renderable, animated
// sprite art. Only species with a working animated source sprite are
// included (base species only — no regional/Mega/Gigantamax forms; see
// cmd/sprite-gen). Sprite artwork is embedded at build time; see /NOTICE
// for attribution.
package sprite

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand/v2"
	"regexp"
	"strings"
	"sync"

	"github.com/GyeongHoKim/pokemon-cli/internal/sprite/spritedata"
)

// Key identifies a species by national Pokédex id. The zero value is never
// a valid key.
type Key int

// ErrUnknownSpecies is returned by Resolve and Load when a name or key does
// not match any known species.
var ErrUnknownSpecies = errors.New("sprite: unknown species")

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// normalizeKey mirrors cmd/sprite-gen's normalizeKey. Duplicated
// intentionally: a 3-line pure function isn't worth exporting just to share
// with the generator.
func normalizeKey(name string) string {
	return nonAlnum.ReplaceAllString(strings.ToLower(name), "")
}

// Resolve looks up a species by name, ignoring case, spaces, hyphens,
// apostrophes, and periods (e.g. "Mr. Mime", "mr-mime", and "MRMIME" all
// resolve to the same Key). Used by cmd/sprite-gen's -preview flag.
func Resolve(name string) (Key, error) {
	key, ok := nameToKey[normalizeKey(name)]
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrUnknownSpecies, name)
	}
	return key, nil
}

var allKeys = sync.OnceValue(func() []Key {
	keys := make([]Key, 0, len(keyToSlug))
	for k := range keyToSlug {
		keys = append(keys, k)
	}
	return keys
})

// Random returns a uniformly random species: its Key and canonical display
// slug (e.g. "charizard", "mr-mime").
func Random() (Key, string) {
	keys := allKeys()
	key := keys[rand.IntN(len(keys))]
	return key, keyToSlug[key]
}

// RandomExcept is like Random but avoids immediately repeating exclude,
// unless the whole pool is a single species.
func RandomExcept(exclude Key) (Key, string) {
	keys := allKeys()
	if len(keys) <= 1 {
		return Random()
	}
	for {
		key, slug := Random()
		if key != exclude {
			return key, slug
		}
	}
}

var loadBundle = sync.OnceValues(func() (spritedata.Bundle, error) {
	gz, err := gzip.NewReader(bytes.NewReader(embeddedData))
	if err != nil {
		return nil, fmt.Errorf("sprite: open embedded data: %w", err)
	}
	defer func() { _ = gz.Close() }()

	var bundle spritedata.Bundle
	if err := gob.NewDecoder(gz).Decode(&bundle); err != nil {
		return nil, fmt.Errorf("sprite: decode embedded data: %w", err)
	}
	return bundle, nil
})

// Sprite holds a resolved species' animation frame sequence.
type Sprite struct {
	frames []spritedata.PixelGrid
}

// Load decodes the embedded sprite bundle (once, cached for the process
// lifetime) and returns the Sprite for key.
func Load(key Key) (*Sprite, error) {
	bundle, err := loadBundle()
	if err != nil {
		return nil, err
	}
	species, ok := bundle[int(key)]
	if !ok {
		return nil, fmt.Errorf("%w: id %d", ErrUnknownSpecies, key)
	}
	return &Sprite{frames: species.Frames}, nil
}

// FrameCount returns how many animation frames this sprite has.
func (s *Sprite) FrameCount() int {
	return len(s.frames)
}

const opaqueAlphaThreshold = 128

// Render returns a ready-to-print, newline-terminated 24-bit truecolor ANSI
// string for the given animation frame (taken mod FrameCount for safety),
// using half-block characters (two source pixel rows per output row).
func (s *Sprite) Render(frame int) string {
	grid := s.frames[frame%len(s.frames)]
	return renderHalfBlocks(grid)
}

func renderHalfBlocks(grid spritedata.PixelGrid) string {
	w, h := int(grid.Width), int(grid.Height)
	var b strings.Builder
	for y := 0; y < h; y += 2 {
		for x := 0; x < w; x++ {
			top := grid.Palette[grid.Indices[y*w+x]]
			topOpaque := top.A >= opaqueAlphaThreshold

			bottom := top
			bottomOpaque := false
			if y+1 < h {
				bottom = grid.Palette[grid.Indices[(y+1)*w+x]]
				bottomOpaque = bottom.A >= opaqueAlphaThreshold
			}

			switch {
			case topOpaque && bottomOpaque:
				fmt.Fprintf(&b, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀", top.R, top.G, top.B, bottom.R, bottom.G, bottom.B)
			case topOpaque:
				fmt.Fprintf(&b, "\x1b[38;2;%d;%d;%dm\x1b[49m▀", top.R, top.G, top.B)
			case bottomOpaque:
				fmt.Fprintf(&b, "\x1b[38;2;%d;%d;%dm\x1b[49m▄", bottom.R, bottom.G, bottom.B)
			default:
				b.WriteString("\x1b[0m ")
			}
		}
		b.WriteString("\x1b[0m\n")
	}
	return b.String()
}
