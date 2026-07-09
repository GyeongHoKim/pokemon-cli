package pet

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/GyeongHoKim/pokemon-cli/internal/sprite"
)

func newTestModel(t *testing.T) Model {
	t.Helper()
	key, err := sprite.Resolve("pikachu")
	if err != nil {
		t.Fatal(err)
	}
	spr, err := sprite.Load(key)
	if err != nil {
		t.Fatal(err)
	}
	return newModelWithSpecies(key, "Pikachu", spr)
}

func TestUpdateWindowSize(t *testing.T) {
	m := newTestModel(t)
	got, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	gm := got.(Model)
	if gm.width != 100 || gm.height != 40 {
		t.Fatalf("got %dx%d, want 100x40", gm.width, gm.height)
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd, got non-nil")
	}
}

func TestUpdateAnimTickAdvancesAndWrapsFrame(t *testing.T) {
	m := newTestModel(t)
	frameCount := m.spr.FrameCount()
	if frameCount < 2 {
		t.Fatalf("test species has only %d frame(s), need at least 2 to test wrap-around", frameCount)
	}
	m.frameIdx = frameCount - 1

	got, cmd := m.Update(animMsg{})
	gm := got.(Model)
	if gm.frameIdx != 0 {
		t.Errorf("frameIdx = %d, want 0 (wrapped)", gm.frameIdx)
	}
	if cmd == nil {
		t.Error("expected a re-arm cmd (next anim tick), got nil")
	}
}

func TestUpdateSwapTickChangesSpeciesAndShowsBubble(t *testing.T) {
	m := newTestModel(t)
	m.frameIdx = 3

	got, cmd := m.Update(swapMsg{})
	gm := got.(Model)
	if gm.key == m.key {
		t.Errorf("key unchanged after swapMsg (RandomExcept should avoid repeats)")
	}
	if gm.displayName == "" {
		t.Error("displayName empty after swap")
	}
	if gm.frameIdx != 0 {
		t.Errorf("frameIdx = %d, want reset to 0 after swap", gm.frameIdx)
	}
	if !gm.bubbleVisible || gm.bubbleText == "" {
		t.Errorf("bubble not shown: visible=%v text=%q", gm.bubbleVisible, gm.bubbleText)
	}
	if cmd == nil {
		t.Error("expected a re-arm cmd (hide tick), got nil")
	}
}

func TestUpdateHideTickHidesBubbleKeepsSpecies(t *testing.T) {
	m := newTestModel(t)
	m.bubbleVisible, m.bubbleText = true, "x"

	got, cmd := m.Update(hideMsg{})
	gm := got.(Model)
	if gm.key != m.key {
		t.Errorf("key changed on hideMsg, want unchanged (species keeps animating)")
	}
	if gm.bubbleVisible || gm.bubbleText != "" {
		t.Errorf("bubble not hidden: visible=%v text=%q", gm.bubbleVisible, gm.bubbleText)
	}
	if cmd == nil {
		t.Error("expected a re-arm cmd (next swap tick), got nil")
	}
}

func TestUpdateQuitKeys(t *testing.T) {
	cases := []tea.KeyPressMsg{
		{Code: 'q'},
		{Code: 'c', Mod: tea.ModCtrl},
	}
	for _, key := range cases {
		m := newTestModel(t)
		_, cmd := m.Update(key)
		if cmd == nil {
			t.Fatalf("key %q: expected non-nil cmd", key.String())
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("key %q: cmd() did not produce tea.QuitMsg", key.String())
		}
	}
}

func TestFormatDisplayName(t *testing.T) {
	cases := map[string]string{
		"pikachu":      "Pikachu",
		"mr-mime":      "Mr Mime",
		"PIKACHU":      "Pikachu",
		"nidoran-f":    "Nidoran F",
		"walking-wake": "Walking Wake",
	}
	for in, want := range cases {
		if got := FormatDisplayName(in); got != want {
			t.Errorf("FormatDisplayName(%q) = %q, want %q", in, got, want)
		}
	}
}
