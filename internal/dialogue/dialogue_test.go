package dialogue

import (
	"strings"
	"testing"
)

func TestArrival(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 200; i++ {
		line := Arrival("Pikachu")
		if line == "" {
			t.Fatal("Arrival returned empty string")
		}
		if !strings.Contains(line, "Pikachu") {
			t.Fatalf("Arrival line %q does not contain display name", line)
		}
		seen[line] = true
	}
	if len(seen) < 2 {
		t.Fatalf("Arrival produced only %d distinct line(s) over 200 calls, want more variety", len(seen))
	}
}
