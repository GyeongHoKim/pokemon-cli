package sprite

import (
	"errors"
	"strings"
	"testing"
)

func TestResolveAndRenderKnownSpecies(t *testing.T) {
	names := []string{
		"pikachu", "charizard", "mr-mime", "nidoran-f", "ho-oh", "walking-wake",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			key, err := Resolve(name)
			if err != nil {
				t.Fatalf("Resolve(%q): %v", name, err)
			}

			s, err := Load(key)
			if err != nil {
				t.Fatalf("Load(%v): %v", key, err)
			}

			if s.FrameCount() == 0 {
				t.Fatal("FrameCount() = 0, want at least 1")
			}
			for frame := 0; frame < s.FrameCount(); frame++ {
				out := s.Render(frame)
				if out == "" {
					t.Fatalf("Render(%d) returned empty string", frame)
				}
				if !strings.Contains(out, "\x1b[") {
					t.Errorf("Render(%d) missing ANSI escape sequences: %q", frame, out)
				}
				if !strings.Contains(out, "\n") {
					t.Errorf("Render(%d) missing newline (not multi-row): %q", frame, out)
				}
			}
		})
	}
}

func TestResolveAliasNormalization(t *testing.T) {
	variants := []string{"Mr. Mime", "MR-MIME", "mrmime", "mr mime"}
	var want Key
	for i, v := range variants {
		got, err := Resolve(v)
		if err != nil {
			t.Fatalf("Resolve(%q): %v", v, err)
		}
		if i == 0 {
			want = got
			continue
		}
		if got != want {
			t.Errorf("Resolve(%q) = %v, want %v (same as Resolve(%q))", v, got, want, variants[0])
		}
	}
}

func TestResolveUnknownSpecies(t *testing.T) {
	_, err := Resolve("not-a-real-pokemon")
	if !errors.Is(err, ErrUnknownSpecies) {
		t.Errorf("Resolve(unknown) error = %v, want ErrUnknownSpecies", err)
	}
}

func TestRegistryCoversAnimatedDex(t *testing.T) {
	const minExpected = 1000 // ~1004/1025 species have an animated source sprite; leave slack
	if len(nameToKey) < minExpected {
		t.Errorf("nameToKey has %d entries, want at least %d (regeneration may be broken)", len(nameToKey), minExpected)
	}
	if len(keyToSlug) != len(nameToKey) {
		t.Errorf("keyToSlug has %d entries, nameToKey has %d; expected them to cover the same species", len(keyToSlug), len(nameToKey))
	}
}

func TestRandom(t *testing.T) {
	key, slug := Random()
	if key == 0 {
		t.Fatal("Random() returned zero Key")
	}
	if slug == "" {
		t.Fatal("Random() returned empty slug")
	}
	if _, err := Load(key); err != nil {
		t.Errorf("Load(Random() key %v): %v", key, err)
	}
}

func TestRandomExceptNeverImmediatelyRepeats(t *testing.T) {
	key, _ := Random()
	for i := 0; i < 200; i++ {
		got, _ := RandomExcept(key)
		if got == key {
			t.Fatalf("RandomExcept(%v) returned the excluded key on attempt %d", key, i)
		}
	}
}
