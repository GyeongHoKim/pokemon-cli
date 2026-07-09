// Package pet implements the bubbletea model for pokemon-cli's terminal
// companion: a random species sits centered in the terminal, animating
// continuously with its real sprite frames, and is periodically replaced by
// a different random species, announced via a speech bubble.
package pet

import (
	"math/rand/v2"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/GyeongHoKim/pokemon-cli/internal/dialogue"
	"github.com/GyeongHoKim/pokemon-cli/internal/sprite"
)

const (
	swapIntervalMin   = 8 * time.Second
	swapIntervalMax   = 20 * time.Second
	bubbleDuration    = 4 * time.Second
	animFrameInterval = 200 * time.Millisecond
)

var bubbleStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(0, 1)

// Model is the bubbletea model driving the terminal pet.
type Model struct {
	key         sprite.Key
	displayName string
	spr         *sprite.Sprite
	frameIdx    int

	bubbleVisible bool
	bubbleText    string

	width, height int
}

// New picks a random species and constructs its initial Model.
func New() (Model, error) {
	key, slug := sprite.Random()
	spr, err := sprite.Load(key)
	if err != nil {
		return Model{}, err
	}
	return Model{key: key, displayName: FormatDisplayName(slug), spr: spr}, nil
}

// newModelWithSpecies builds a Model for a specific, already-loaded species
// — used by tests that need deterministic (non-random) state.
func newModelWithSpecies(key sprite.Key, displayName string, spr *sprite.Sprite) Model {
	return Model{key: key, displayName: displayName, spr: spr}
}

// FormatDisplayName title-cases a canonical species slug for use in dialogue
// text (e.g. "mr-mime" -> "Mr Mime").
func FormatDisplayName(raw string) string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, " ")
}

// randDuration returns a random non-negative duration less than max.
func randDuration(max time.Duration) time.Duration {
	return time.Duration(rand.Int64N(int64(max)))
}

// animMsg drives the continuous per-species frame animation; it re-arms
// itself every tick regardless of any other state. swapMsg and hideMsg form
// a separate, independent self-re-arming ping-pong timer: swapMsg picks a
// new random species and shows an arrival bubble, arming hideMsg; hideMsg
// hides the bubble (species keeps animating) and arms the next swapMsg.
// Exactly one swap/hide timer is ever in flight, so there's no stale-tick
// race to guard against there.
type animMsg struct{}
type swapMsg struct{}
type hideMsg struct{}

func scheduleAnim() tea.Cmd {
	return tea.Tick(animFrameInterval, func(time.Time) tea.Msg { return animMsg{} })
}

func scheduleSwap() tea.Cmd {
	d := swapIntervalMin + randDuration(swapIntervalMax-swapIntervalMin)
	return tea.Tick(d, func(time.Time) tea.Msg { return swapMsg{} })
}

func scheduleHide() tea.Cmd {
	return tea.Tick(bubbleDuration, func(time.Time) tea.Msg { return hideMsg{} })
}

// Init starts the continuous animation timer and the species-swap timer.
func (m Model) Init() tea.Cmd {
	return tea.Batch(scheduleAnim(), scheduleSwap())
}

// Update handles window resizes, frame animation, the swap/hide timer
// ping-pong, and quit keys.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case animMsg:
		m.frameIdx = (m.frameIdx + 1) % m.spr.FrameCount()
		return m, scheduleAnim()

	case swapMsg:
		key, slug := sprite.RandomExcept(m.key)
		if spr, err := sprite.Load(key); err == nil {
			// The embedded data is self-consistent (registry and bundle are
			// generated together), so Load failing here should never happen
			// in practice. Skip the swap rather than crashing a running program.
			m.key, m.displayName, m.spr, m.frameIdx = key, FormatDisplayName(slug), spr, 0
		}
		m.bubbleVisible = true
		m.bubbleText = dialogue.Arrival(m.displayName)
		return m, scheduleHide()

	case hideMsg:
		m.bubbleVisible = false
		m.bubbleText = ""
		return m, scheduleSwap()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the current animation frame (and, when a new species just
// arrived, the speech bubble) centered in the terminal.
func (m Model) View() tea.View {
	block := strings.TrimRight(m.spr.Render(m.frameIdx), "\n")

	content := block
	if m.bubbleVisible {
		content = lipgloss.JoinVertical(lipgloss.Center, block, bubbleStyle.Render(m.bubbleText))
	}

	w, h := m.width, m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}

	v := tea.NewView(lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content))
	v.AltScreen = true
	return v
}
