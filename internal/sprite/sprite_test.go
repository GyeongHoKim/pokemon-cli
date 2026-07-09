package sprite

import (
	"errors"
	"strings"
	"testing"
)

func TestResolveAndRenderKnownSpecies(t *testing.T) {
	names := []string{
		"pikachu", "charizard", "mr-mime", "nidoran-f", "ho-oh", "ogerpon", "walking-wake",
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

			for _, pose := range []Pose{PoseIcon, PoseBattle} {
				out := s.Render(pose)
				if out == "" {
					t.Fatalf("Render(pose=%v) returned empty string", pose)
				}
				if !strings.Contains(out, "\x1b[") {
					t.Errorf("Render(pose=%v) missing ANSI escape sequences: %q", pose, out)
				}
				if !strings.Contains(out, "\n") {
					t.Errorf("Render(pose=%v) missing newline (not multi-row): %q", pose, out)
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

func TestRegistryCoversFullDex(t *testing.T) {
	const minExpected = 1000 // full national dex is 1025; leave slack for future gens
	if len(nameToKey) < minExpected {
		t.Errorf("nameToKey has %d entries, want at least %d (regeneration may be broken)", len(nameToKey), minExpected)
	}
}
