// Command pokemon-cli shows a randomly chosen Pokémon species as a terminal
// companion: it sits centered in the terminal animating continuously, and
// is periodically replaced by a different random species, announced via a
// speech bubble.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/GyeongHoKim/pokemon-cli/internal/pet"
)

func main() {
	m, err := pet.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "pokemon-cli: %v\n", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(m, tea.WithEnvironment(syncOutputSafeEnviron())).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "pokemon-cli: %v\n", err)
		os.Exit(1)
	}
}

// syncOutputSafeEnviron returns the process environment with WT_SESSION and
// TERM_PROGRAM stripped and SSH_TTY stubbed in, so bubbletea's terminal
// capability probe never queries for synchronized-output (mode 2026)
// support. Real TERM/COLORTERM values are left untouched so color-profile
// detection is unaffected. Windows Terminal (common under WSL2) always sets
// WT_SESSION, which otherwise unconditionally triggers that query — and a
// terminal that answers it can leave bubbletea v2's renderer stuck
// re-painting only the first frame (observed on WSL2 + Windows Terminal;
// tracked upstream in charmbracelet/bubbletea's renderer flush/resize
// handling). Dropping the query sidesteps the bug entirely.
func syncOutputSafeEnviron() []string {
	environ := os.Environ()
	filtered := make([]string, 0, len(environ)+1)
	for _, kv := range environ {
		if strings.HasPrefix(kv, "WT_SESSION=") || strings.HasPrefix(kv, "TERM_PROGRAM=") {
			continue
		}
		filtered = append(filtered, kv)
	}
	return append(filtered, "SSH_TTY=/dev/pts/0")
}
